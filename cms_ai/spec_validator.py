from __future__ import annotations

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class ValidationError:
    path: str
    message: str


def validate_template_spec(spec: dict[str, Any]) -> list[ValidationError]:
    """Validate a minimal Template Spec shape.

    This is intentionally small: it lets us start writing tests and
    iteratively expand toward the full spec.
    """

    errors: list[ValidationError] = []

    if not isinstance(spec, dict):
        return [ValidationError(path="$", message="spec must be an object")]

    tokens = spec.get("tokens")
    layouts = spec.get("layouts")

    if not isinstance(tokens, dict):
        errors.append(
            ValidationError(path="$.tokens", message="tokens must be an object")
        )

    if not isinstance(layouts, list) or not layouts:
        errors.append(
            ValidationError(
                path="$.layouts", message="layouts must be a non-empty array"
            )
        )
        return errors

    constraints_raw = spec.get("constraints")
    constraints: dict[str, Any]
    if isinstance(constraints_raw, dict):
        constraints = constraints_raw
    else:
        constraints = {}

    safe_margin = constraints.get("safeMargin", 0.05)

    if (
        not isinstance(safe_margin, (int, float))
        or safe_margin < 0
        or safe_margin >= 0.5
    ):
        errors.append(
            ValidationError(
                path="$.constraints.safeMargin",
                message="safeMargin must be a number in [0, 0.5)",
            )
        )
        safe_margin = 0.05

    for layout_index, layout in enumerate(layouts):
        layout_path = f"$.layouts[{layout_index}]"

        if not isinstance(layout, dict):
            errors.append(
                ValidationError(path=layout_path, message="layout must be an object")
            )
            continue

        name = layout.get("name")
        if not isinstance(name, str) or not name.strip():
            errors.append(
                ValidationError(path=f"{layout_path}.name", message="name is required")
            )

        placeholders = layout.get("placeholders")
        if not isinstance(placeholders, list) or not placeholders:
            errors.append(
                ValidationError(
                    path=f"{layout_path}.placeholders",
                    message="placeholders must be a non-empty array",
                )
            )
            continue

        rects: list[tuple[float, float, float, float, str]] = []

        for placeholder_index, placeholder in enumerate(placeholders):
            placeholder_path = f"{layout_path}.placeholders[{placeholder_index}]"

            if not isinstance(placeholder, dict):
                errors.append(
                    ValidationError(
                        path=placeholder_path, message="placeholder must be an object"
                    )
                )
                continue

            placeholder_id = placeholder.get("id")
            if not isinstance(placeholder_id, str) or not placeholder_id.strip():
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.id", message="id is required"
                    )
                )
                placeholder_id = f"{layout_index}:{placeholder_index}"

            geometry = placeholder.get("geometry")
            if not isinstance(geometry, dict):
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry",
                        message="geometry must be an object with x/y/w/h",
                    )
                )
                continue

            x = geometry.get("x")
            y = geometry.get("y")
            w = geometry.get("w")
            h = geometry.get("h")

            if not isinstance(x, (int, float)):
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry.x",
                        message="x must be a number",
                    )
                )
                continue
            if not isinstance(y, (int, float)):
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry.y",
                        message="y must be a number",
                    )
                )
                continue
            if not isinstance(w, (int, float)):
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry.w",
                        message="w must be a number",
                    )
                )
                continue
            if not isinstance(h, (int, float)):
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry.h",
                        message="h must be a number",
                    )
                )
                continue

            x_f = float(x)
            y_f = float(y)
            w_f = float(w)
            h_f = float(h)

            if w_f <= 0 or h_f <= 0:
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry",
                        message="w and h must be > 0",
                    )
                )
                continue

            if x_f < safe_margin or y_f < safe_margin:
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry",
                        message="x/y must respect safe margins",
                    )
                )

            if x_f + w_f > 1.0 - safe_margin or y_f + h_f > 1.0 - safe_margin:
                errors.append(
                    ValidationError(
                        path=f"{placeholder_path}.geometry",
                        message="geometry must fit within safe margins",
                    )
                )

            rects.append((x_f, y_f, w_f, h_f, str(placeholder_id)))

        for i in range(len(rects)):
            ax, ay, aw, ah, aid = rects[i]
            for j in range(i + 1, len(rects)):
                bx, by, bw, bh, bid = rects[j]
                if _rects_overlap(ax, ay, aw, ah, bx, by, bw, bh):
                    errors.append(
                        ValidationError(
                            path=layout_path,
                            message=f"placeholders overlap: {aid} and {bid}",
                        )
                    )

    return errors


def _rects_overlap(
    ax: float,
    ay: float,
    aw: float,
    ah: float,
    bx: float,
    by: float,
    bw: float,
    bh: float,
) -> bool:
    # Treat touching edges as non-overlapping.
    if ax + aw <= bx or bx + bw <= ax:
        return False
    if ay + ah <= by or by + bh <= ay:
        return False
    return True
