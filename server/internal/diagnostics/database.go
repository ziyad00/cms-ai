package diagnostics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ziyad/cms-ai/server/internal/logger"
)

type DatabaseDiagnostics struct {
	db *sql.DB
}

func NewDatabaseDiagnostics(db *sql.DB) *DatabaseDiagnostics {
	return &DatabaseDiagnostics{db: db}
}

type TemplateAnalysis struct {
	TotalTemplates       int `json:"total_templates"`
	TemplatesWithVersion int `json:"templates_with_version"`
	TemplatesWithoutVersion int `json:"templates_without_version"`
	OrphanedTemplates    int `json:"orphaned_templates"`
}

type OrganizationStats struct {
	OrgID                   string `json:"org_id"`
	TotalTemplates          int    `json:"total_templates"`
	TemplatesWithoutVersion int    `json:"templates_without_version"`
	HasProblematicData      bool   `json:"has_problematic_data"`
}

func (d *DatabaseDiagnostics) AnalyzeTemplates(ctx context.Context) (*TemplateAnalysis, error) {
	logger.Database().Info("starting_template_analysis")

	// Get total templates
	totalQuery := `SELECT COUNT(*) FROM templates`
	var total int
	err := d.db.QueryRowContext(ctx, totalQuery).Scan(&total)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "count_total_templates", err)
		return nil, err
	}

	// Get templates with versions
	withVersionQuery := `SELECT COUNT(*) FROM templates WHERE current_version_id IS NOT NULL`
	var withVersion int
	err = d.db.QueryRowContext(ctx, withVersionQuery).Scan(&withVersion)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "count_with_version", err)
		return nil, err
	}

	// Get orphaned templates (templates without any versions)
	orphanedQuery := `
		SELECT COUNT(*) FROM templates t
		WHERE NOT EXISTS (
			SELECT 1 FROM template_versions tv
			WHERE tv.template_id = t.id
		)`
	var orphaned int
	err = d.db.QueryRowContext(ctx, orphanedQuery).Scan(&orphaned)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "count_orphaned", err)
		return nil, err
	}

	analysis := &TemplateAnalysis{
		TotalTemplates:          total,
		TemplatesWithVersion:    withVersion,
		TemplatesWithoutVersion: total - withVersion,
		OrphanedTemplates:       orphaned,
	}

	logger.Database().Info("template_analysis_completed",
		"total", total,
		"with_version", withVersion,
		"without_version", total-withVersion,
		"orphaned", orphaned,
	)

	return analysis, nil
}

func (d *DatabaseDiagnostics) AnalyzeByOrganization(ctx context.Context) ([]OrganizationStats, error) {
	logger.Database().Info("starting_organization_analysis")

	query := `
		SELECT
			o.id as org_id,
			COUNT(t.id) as total_templates,
			COUNT(CASE WHEN t.current_version_id IS NULL THEN 1 END) as without_version
		FROM organizations o
		LEFT JOIN templates t ON t.org_id = o.id
		GROUP BY o.id
		HAVING COUNT(t.id) > 0
		ORDER BY without_version DESC, total_templates DESC`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "analyze_by_organization", err)
		return nil, err
	}
	defer rows.Close()

	var stats []OrganizationStats
	for rows.Next() {
		var stat OrganizationStats
		err := rows.Scan(&stat.OrgID, &stat.TotalTemplates, &stat.TemplatesWithoutVersion)
		if err != nil {
			logger.LogError(ctx, "diagnostics", "scan_organization_stats", err)
			continue
		}

		stat.HasProblematicData = stat.TemplatesWithoutVersion > 0
		stats = append(stats, stat)
	}

	logger.Database().Info("organization_analysis_completed",
		"organizations_analyzed", len(stats),
	)

	return stats, nil
}

func (d *DatabaseDiagnostics) GetProblematicTemplates(ctx context.Context, limit int) ([]ProblematicTemplate, error) {
	logger.Database().Info("fetching_problematic_templates", "limit", limit)

	query := `
		SELECT
			t.id,
			t.org_id,
			t.name,
			t.status,
			t.current_version_id,
			t.latest_version_no,
			t.created_at,
			CASE WHEN tv.id IS NOT NULL THEN true ELSE false END as has_versions
		FROM templates t
		LEFT JOIN template_versions tv ON tv.template_id = t.id
		WHERE t.current_version_id IS NULL
		GROUP BY t.id, t.org_id, t.name, t.status, t.current_version_id, t.latest_version_no, t.created_at
		ORDER BY t.created_at DESC
		LIMIT $1`

	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		logger.LogError(ctx, "diagnostics", "fetch_problematic_templates", err)
		return nil, err
	}
	defer rows.Close()

	var templates []ProblematicTemplate
	for rows.Next() {
		var template ProblematicTemplate
		var currentVersionID sql.NullString
		err := rows.Scan(
			&template.ID,
			&template.OrgID,
			&template.Name,
			&template.Status,
			&currentVersionID,
			&template.LatestVersionNo,
			&template.CreatedAt,
			&template.HasVersions,
		)
		if err != nil {
			logger.LogError(ctx, "diagnostics", "scan_problematic_template", err)
			continue
		}

		if currentVersionID.Valid {
			template.CurrentVersionID = &currentVersionID.String
		}

		templates = append(templates, template)
	}

	logger.Database().Info("problematic_templates_fetched",
		"count", len(templates),
	)

	return templates, nil
}

type ProblematicTemplate struct {
	ID               string    `json:"id"`
	OrgID            string    `json:"org_id"`
	Name             string    `json:"name"`
	Status           string    `json:"status"`
	CurrentVersionID *string   `json:"current_version_id"`
	LatestVersionNo  int       `json:"latest_version_no"`
	CreatedAt        time.Time `json:"created_at"`
	HasVersions      bool      `json:"has_versions"`
}

func (d *DatabaseDiagnostics) FixOrphanedTemplates(ctx context.Context, dryRun bool) (int, error) {
	operation := "fix_orphaned_templates"
	if dryRun {
		operation = "dry_run_fix_orphaned_templates"
	}

	logger.Database().Info("starting_orphan_fix", "dry_run", dryRun)

	// Find templates that have versions but current_version_id is NULL
	query := `
		SELECT DISTINCT t.id, tv.id as version_id
		FROM templates t
		INNER JOIN template_versions tv ON tv.template_id = t.id
		WHERE t.current_version_id IS NULL
		ORDER BY t.id, tv.created_at ASC`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		logger.LogError(ctx, "diagnostics", operation, err)
		return 0, err
	}
	defer rows.Close()

	var fixes []struct {
		TemplateID string
		VersionID  string
	}

	for rows.Next() {
		var fix struct {
			TemplateID string
			VersionID  string
		}
		err := rows.Scan(&fix.TemplateID, &fix.VersionID)
		if err != nil {
			continue
		}
		fixes = append(fixes, fix)
	}

	if dryRun {
		logger.Database().Info("dry_run_orphan_fix_completed",
			"templates_to_fix", len(fixes),
		)
		return len(fixes), nil
	}

	// Apply fixes
	updateQuery := `UPDATE templates SET current_version_id = $1 WHERE id = $2`
	fixed := 0

	for _, fix := range fixes {
		_, err := d.db.ExecContext(ctx, updateQuery, fix.VersionID, fix.TemplateID)
		if err != nil {
			logger.LogError(ctx, "diagnostics", "apply_fix", err,
				"template_id", fix.TemplateID,
				"version_id", fix.VersionID,
			)
			continue
		}
		fixed++
	}

	logger.Database().Info("orphan_fix_completed",
		"templates_fixed", fixed,
		"templates_attempted", len(fixes),
	)

	return fixed, nil
}

func (d *DatabaseDiagnostics) RunFullDiagnostics(ctx context.Context) error {
	logger.Database().Info("starting_full_database_diagnostics")

	// Template analysis
	analysis, err := d.AnalyzeTemplates(ctx)
	if err != nil {
		return fmt.Errorf("template analysis failed: %w", err)
	}

	// Organization analysis
	orgStats, err := d.AnalyzeByOrganization(ctx)
	if err != nil {
		return fmt.Errorf("organization analysis failed: %w", err)
	}

	// Get problematic templates
	problematic, err := d.GetProblematicTemplates(ctx, 10)
	if err != nil {
		return fmt.Errorf("problematic templates fetch failed: %w", err)
	}

	// Log comprehensive summary
	logger.Database().Info("full_diagnostics_completed",
		"total_templates", analysis.TotalTemplates,
		"templates_without_version", analysis.TemplatesWithoutVersion,
		"orphaned_templates", analysis.OrphanedTemplates,
		"organizations_with_issues", countProblematicOrgs(orgStats),
		"sample_problematic_templates", len(problematic),
	)

	return nil
}

func countProblematicOrgs(stats []OrganizationStats) int {
	count := 0
	for _, stat := range stats {
		if stat.HasProblematicData {
			count++
		}
	}
	return count
}