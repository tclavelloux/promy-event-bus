#!/usr/bin/env bash
set -euo pipefail

ERRORS=0

error() {
  echo "ERROR: $1"
  ERRORS=$((ERRORS + 1))
}

VALID_TYPES="string number boolean object array"
VALID_FORMATS="uuid email date-time uri"

# Validate stream directories
for stream_dir in registry/streams/*/; do
  if [ ! -f "${stream_dir}stream.yaml" ]; then
    error "$stream_dir -- missing stream.yaml"
  fi
done

# Validate event files
for event_file in registry/streams/*/events/*.yaml; do
  [ -f "$event_file" ] || continue

  filename=$(basename "$event_file" .yaml)

  # Rule: stream.yaml exists in parent directory
  stream_dir=$(dirname "$(dirname "$event_file")")
  if [ ! -f "${stream_dir}/stream.yaml" ]; then
    error "$event_file -- no stream.yaml in parent directory"
  fi

  # Rule: name matches filename
  declared_name=$(yq -r '.name' "$event_file")
  if [ "$declared_name" != "$filename" ]; then
    error "$event_file -- 'name: $declared_name' does not match filename '$filename'"
  fi

  # Rule: event name is dot-separated snake_case segments (e.g. user.registered, user.preferences.updated)
  if ! echo "$filename" | grep -qE '^[a-z]+(_[a-z]+)*(\.[a-z]+(_[a-z]+)*)+$'; then
    error "$event_file -- event name '$filename' must be dot-separated snake_case (e.g. user.registered)"
  fi

  # Rule: tier is 1 or 2
  tier=$(yq -r '.tier' "$event_file")
  if [[ "$tier" != "1" && "$tier" != "2" ]]; then
    error "$event_file -- 'tier' must be 1 or 2, got '$tier'"
  fi

  # Rule: description present and non-empty
  description=$(yq -r '.description' "$event_file")
  if [ -z "$description" ] || [ "$description" = "null" ]; then
    error "$event_file -- 'description' is missing or empty"
  fi

  # Rule: fields present and non-empty
  fields_count=$(yq -r '.fields | length' "$event_file" 2>/dev/null || echo "0")
  if [ "$fields_count" = "0" ] || [ "$fields_count" = "null" ]; then
    error "$event_file -- 'fields' is missing or empty"
  fi

  # Rule: example present
  example=$(yq -r '.example' "$event_file")
  if [ -z "$example" ] || [ "$example" = "null" ]; then
    error "$event_file -- 'example' is missing"
  fi

  # Validate each field
  field_names=$(yq -r '.fields | keys | .[]' "$event_file" 2>/dev/null || true)
  for field in $field_names; do
    # Rule: field name is snake_case
    if ! echo "$field" | grep -qE '^[a-z]+(_[a-z]+)*$'; then
      error "$event_file -- field '$field' is not snake_case"
    fi

    # Rule: field has type
    field_type=$(yq -r ".fields.\"$field\".type" "$event_file")
    if [ -z "$field_type" ] || [ "$field_type" = "null" ]; then
      error "$event_file -- field '$field' is missing 'type'"
    else
      # Rule: type is a valid JSON Schema type
      if ! echo "$VALID_TYPES" | grep -qw "$field_type"; then
        error "$event_file -- field '$field' has invalid type '$field_type' (allowed: $VALID_TYPES)"
      fi
    fi

    # Rule: field has required
    field_required=$(yq -r ".fields.\"$field\".required" "$event_file")
    if [ -z "$field_required" ] || [ "$field_required" = "null" ]; then
      error "$event_file -- field '$field' is missing 'required'"
    fi

    # Rule: field has description
    field_desc=$(yq -r ".fields.\"$field\".description" "$event_file")
    if [ -z "$field_desc" ] || [ "$field_desc" = "null" ]; then
      error "$event_file -- field '$field' is missing 'description'"
    fi

    # Rule: format (if present) is from controlled vocabulary
    field_format=$(yq -r ".fields.\"$field\".format" "$event_file")
    if [ -n "$field_format" ] && [ "$field_format" != "null" ]; then
      if ! echo "$VALID_FORMATS" | grep -qw "$field_format"; then
        error "$event_file -- field '$field' has invalid format '$field_format' (allowed: $VALID_FORMATS)"
      fi
    fi
  done
done

echo ""
if [ "$ERRORS" -gt 0 ]; then
  echo "Registry validation failed with $ERRORS error(s). Fix the above before merging."
  exit 1
fi

echo "Registry validation passed."
