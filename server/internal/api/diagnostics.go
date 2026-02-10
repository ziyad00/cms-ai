package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/ziyad/cms-ai/server/internal/diagnostics"
	"github.com/ziyad/cms-ai/server/internal/logger"
	"github.com/ziyad/cms-ai/server/internal/store/postgres"
)

func (s *Server) handleDatabaseDiagnostics(w http.ResponseWriter, r *http.Request) {
	// Get PostgreSQL database from store
	pgStore, ok := s.Store.(*postgres.PostgresStore)
	if !ok {
		writeError(w, r, http.StatusInternalServerError, "database diagnostics only available for PostgreSQL")
		return
	}

	// Access the underlying database connection
	db, err := pgStore.DB()

	if err != nil || db == nil {
		writeError(w, r, http.StatusInternalServerError, "database connection not available")
		return
	}

	diag := diagnostics.NewDatabaseDiagnostics(db)

	ctx := r.Context()
	logger.API().Info("running_database_diagnostics")

	// Run full diagnostics
	err = diag.RunFullDiagnostics(ctx)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "run_full_diagnostics", err)
		writeError(w, r, http.StatusInternalServerError, "diagnostics failed")
		return
	}

	// Get detailed analysis
	analysis, err := diag.AnalyzeTemplates(ctx)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "analyze_templates", err)
		writeError(w, r, http.StatusInternalServerError, "template analysis failed")
		return
	}

	orgStats, err := diag.AnalyzeByOrganization(ctx)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "analyze_by_organization", err)
		writeError(w, r, http.StatusInternalServerError, "organization analysis failed")
		return
	}

	problematic, err := diag.GetProblematicTemplates(ctx, 20)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "get_problematic_templates", err)
		writeError(w, r, http.StatusInternalServerError, "problematic templates fetch failed")
		return
	}

	response := map[string]interface{}{
		"analysis":              analysis,
		"organization_stats":    orgStats,
		"problematic_templates": problematic,
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleDatabaseQuery(w http.ResponseWriter, r *http.Request) {
	// Security check - only allow specific safe queries
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, r, http.StatusBadRequest, "query parameter required")
		return
	}

	// Get PostgreSQL database
	pgStore, ok := s.Store.(*postgres.PostgresStore)
	if !ok {
		writeError(w, r, http.StatusInternalServerError, "raw queries only available for PostgreSQL")
		return
	}

	db, err := pgStore.DB()
	if err != nil || db == nil {
		writeError(w, r, http.StatusInternalServerError, "database connection not available")
		return
	}

	ctx := r.Context()
	logger.Database().Info("executing_raw_query", "query", query)

	// Execute query based on predefined safe queries
	var result interface{}

	switch query {
	case "template_stats":
		result, err = s.queryTemplateStats(ctx, db)
	case "template_versions_stats":
		result, err = s.queryTemplateVersionStats(ctx, db)
	case "empty_specs":
		result, err = s.queryEmptySpecs(ctx, db)
	case "null_current_version":
		result, err = s.queryNullCurrentVersion(ctx, db)
	case "sample_templates":
		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		result, err = s.querySampleTemplates(ctx, db, limit)
	case "sample_versions":
		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		result, err = s.querySampleVersions(ctx, db, limit)
	case "organization_list":
		result, err = s.queryOrganizations(ctx, db)
	default:
		writeError(w, r, http.StatusBadRequest, "unsupported query. Available: template_stats, template_versions_stats, empty_specs, null_current_version, sample_templates, sample_versions, organization_list")
		return
	}

	if err != nil {
		logger.LogError(ctx, "diagnostics", "raw_query", err, "query", query)
		writeError(w, r, http.StatusInternalServerError, "query failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":  query,
		"result": result,
	})
}

// Safe predefined queries
func (s *Server) queryTemplateStats(ctx context.Context, dbInterface interface{}) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			COUNT(*) as total_templates,
			COUNT(CASE WHEN current_version_id IS NOT NULL THEN 1 END) as with_version,
			COUNT(CASE WHEN current_version_id IS NULL THEN 1 END) as without_version
		FROM templates`

	var stats struct {
		Total       int `json:"total_templates"`
		WithVersion int `json:"with_version"`
		WithoutVersion int `json:"without_version"`
	}

	err := db.QueryRowContext(ctx, query).Scan(&stats.Total, &stats.WithVersion, &stats.WithoutVersion)
	return stats, err
}

func (s *Server) queryTemplateVersionStats(ctx context.Context, dbInterface interface{}) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			COUNT(*) as total_versions,
			COUNT(CASE WHEN spec_json IS NULL THEN 1 END) as null_specs,
			COUNT(CASE WHEN spec_json = '{}' THEN 1 END) as empty_specs,
			COUNT(CASE WHEN spec_json IS NOT NULL AND spec_json != '{}' THEN 1 END) as valid_specs
		FROM template_versions`

	var stats struct {
		Total      int `json:"total_versions"`
		NullSpecs  int `json:"null_specs"`
		EmptySpecs int `json:"empty_specs"`
		ValidSpecs int `json:"valid_specs"`
	}

	err := db.QueryRowContext(ctx, query).Scan(&stats.Total, &stats.NullSpecs, &stats.EmptySpecs, &stats.ValidSpecs)
	return stats, err
}

func (s *Server) queryEmptySpecs(ctx context.Context, dbInterface interface{}) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			tv.id,
			tv.template_id,
			tv.version_no,
			tv.spec_json,
			tv.created_at,
			t.name as template_name,
			t.org_id
		FROM template_versions tv
		JOIN templates t ON t.id = tv.template_id
		WHERE tv.spec_json IS NULL OR tv.spec_json = '{}' OR tv.spec_json = 'null'
		ORDER BY tv.created_at DESC
		LIMIT 20`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, templateID, templateName, orgID string
		var versionNo int
		var specJSON sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &templateID, &versionNo, &specJSON, &createdAt, &templateName, &orgID)
		if err != nil {
			continue
		}

		var spec interface{}
		if specJSON.Valid {
			spec = specJSON.String
		}

		results = append(results, map[string]interface{}{
			"id":            id,
			"template_id":   templateID,
			"template_name": templateName,
			"org_id":        orgID,
			"version_no":    versionNo,
			"spec_json":     spec,
			"created_at":    createdAt,
		})
	}

	return results, nil
}

func (s *Server) queryNullCurrentVersion(ctx context.Context, dbInterface interface{}) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			t.id,
			t.name,
			t.org_id,
			t.status,
			t.latest_version_no,
			t.created_at,
			COUNT(tv.id) as version_count
		FROM templates t
		LEFT JOIN template_versions tv ON tv.template_id = t.id
		WHERE t.current_version_id IS NULL
		GROUP BY t.id, t.name, t.org_id, t.status, t.latest_version_no, t.created_at
		ORDER BY t.created_at DESC
		LIMIT 20`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name, orgID, status string
		var latestVersionNo, versionCount int
		var createdAt time.Time

		err := rows.Scan(&id, &name, &orgID, &status, &latestVersionNo, &createdAt, &versionCount)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":                id,
			"name":              name,
			"org_id":            orgID,
			"status":            status,
			"latest_version_no": latestVersionNo,
			"version_count":     versionCount,
			"created_at":        createdAt,
		})
	}

	return results, nil
}

func (s *Server) querySampleTemplates(ctx context.Context, dbInterface interface{}, limit int) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			id, name, org_id, status, current_version_id, latest_version_no, created_at
		FROM templates
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name, orgID, status string
		var currentVersionID sql.NullString
		var latestVersionNo int
		var createdAt time.Time

		err := rows.Scan(&id, &name, &orgID, &status, &currentVersionID, &latestVersionNo, &createdAt)
		if err != nil {
			continue
		}

		var currentVer interface{}
		if currentVersionID.Valid {
			currentVer = currentVersionID.String
		}

		results = append(results, map[string]interface{}{
			"id":                 id,
			"name":               name,
			"org_id":             orgID,
			"status":             status,
			"current_version_id": currentVer,
			"latest_version_no":  latestVersionNo,
			"created_at":         createdAt,
		})
	}

	return results, nil
}

func (s *Server) querySampleVersions(ctx context.Context, dbInterface interface{}, limit int) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			tv.id, tv.template_id, tv.version_no, tv.spec_json, tv.created_at,
			t.name as template_name, t.org_id
		FROM template_versions tv
		JOIN templates t ON t.id = tv.template_id
		ORDER BY tv.created_at DESC
		LIMIT $1`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, templateID, templateName, orgID string
		var versionNo int
		var specJSON sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &templateID, &versionNo, &specJSON, &createdAt, &templateName, &orgID)
		if err != nil {
			continue
		}

		var spec interface{}
		if specJSON.Valid {
			spec = specJSON.String
		}

		results = append(results, map[string]interface{}{
			"id":            id,
			"template_id":   templateID,
			"template_name": templateName,
			"org_id":        orgID,
			"version_no":    versionNo,
			"spec_json":     spec,
			"created_at":    createdAt,
		})
	}

	return results, nil
}

func (s *Server) queryOrganizations(ctx context.Context, dbInterface interface{}) (interface{}, error) {
	db := dbInterface.(*sql.DB)
	query := `
		SELECT
			o.id, o.name, o.created_at,
			COUNT(t.id) as template_count,
			COUNT(CASE WHEN t.current_version_id IS NULL THEN 1 END) as templates_without_version
		FROM organizations o
		LEFT JOIN templates t ON t.org_id = o.id
		GROUP BY o.id, o.name, o.created_at
		ORDER BY template_count DESC, o.created_at DESC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name string
		var createdAt time.Time
		var templateCount, templatesWithoutVersion int

		err := rows.Scan(&id, &name, &createdAt, &templateCount, &templatesWithoutVersion)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":                        id,
			"name":                      name,
			"created_at":                createdAt,
			"template_count":            templateCount,
			"templates_without_version": templatesWithoutVersion,
		})
	}

	return results, nil
}