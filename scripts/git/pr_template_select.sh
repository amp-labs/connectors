#!/usr/bin/env bash

## Main variables
TEMPLATE_DIR=".github/PULL_REQUEST_TEMPLATE"
TEMP_PR_BODY=".git/.pr_body.tmp"
TEMP_PR_TITLE=".git/.pr_title.tmp"

##############################
# Display PR template options.
##############################
declare -a TEMPLATE_FILES
declare -a TEMPLATE_NAMES
declare -a TEMPLATE_PRIORITIES

TEMPLATE_DIR=".github/PULL_REQUEST_TEMPLATE"

for FILE in "$TEMPLATE_DIR"/*.md; do
  # Extract template_name
  NAME=$(sed -n 's/^template_name:[[:space:]]*"\?\([^"]*\)"\?/\1/p' "$FILE")
  [[ -z "$NAME" ]] && NAME=$(basename "$FILE" .md | sed -E 's/_/ /g; s/(^| )([a-z])/\U\2/g')

  # Extract priority (default 100)
  PRIORITY=$(sed -n 's/^priority:[[:space:]]*\([0-9]*\)/\1/p' "$FILE")
  [[ -z "$PRIORITY" ]] && PRIORITY=100

  TEMPLATE_FILES+=("$FILE")
  TEMPLATE_NAMES+=("$NAME")
  TEMPLATE_PRIORITIES+=("$PRIORITY")
done

# Append None and Cancel
TEMPLATE_FILES+=("")
TEMPLATE_NAMES+=("[None]")
TEMPLATE_PRIORITIES+=(999)
TEMPLATE_FILES+=("")
TEMPLATE_NAMES+=("[Cancel]")
TEMPLATE_PRIORITIES+=(1000)

# Sort by priority (simple bubble sort for small arrays)
for ((i=0;i<${#TEMPLATE_FILES[@]}-1;i++)); do
  for ((j=i+1;j<${#TEMPLATE_FILES[@]};j++)); do
    if (( TEMPLATE_PRIORITIES[i] > TEMPLATE_PRIORITIES[j] )); then
      # Swap files
      tmp="${TEMPLATE_FILES[i]}"; TEMPLATE_FILES[i]="${TEMPLATE_FILES[j]}"; TEMPLATE_FILES[j]="$tmp"
      # Swap names
      tmp="${TEMPLATE_NAMES[i]}"; TEMPLATE_NAMES[i]="${TEMPLATE_NAMES[j]}"; TEMPLATE_NAMES[j]="$tmp"
      # Swap priorities
      tmp="${TEMPLATE_PRIORITIES[i]}"; TEMPLATE_PRIORITIES[i]="${TEMPLATE_PRIORITIES[j]}"; TEMPLATE_PRIORITIES[j]="$tmp"
    fi
  done
done

# Display menu
echo "Available PR Templates:"
select TEMPLATE_SELECTION in "${TEMPLATE_NAMES[@]}"; do
  # User pressed Enter without selection
  if [[ -z "$REPLY" || "$REPLY" -gt ${#TEMPLATE_NAMES[@]} ]]; then
    echo "No template selected → using None"
    TEMPLATE_FILE=""
    break
  fi

  INDEX=$((REPLY-1))
  CHOICE="${TEMPLATE_NAMES[INDEX]}"
  TEMPLATE_FILE="${TEMPLATE_FILES[INDEX]}"

  case "$CHOICE" in
    "[None]")
      echo "You selected: None"
      TEMPLATE_FILE=""
      exit 0
      ;;
    "[Cancel]")
      echo "Cancelled."
      exit 1
      ;;
    *)
      echo "You selected: $CHOICE file($TEMPLATE_FILE)"
      break
      ;;
  esac
done < /dev/tty

##############################
# Extract YAML metadata at the top of the file
##############################
METADATA=$(sed -n '/^---/,/^---/p' "$TEMPLATE_FILE" | sed '1d;$d')
PR_TITLE=$(echo "$METADATA" | sed -n 's/^pr_title:[[:space:]]*"\?\([^"]*\)"\?/\1/p')
# Parse dynamic fields
FIELDS=$(echo "$METADATA" | sed -n '/fields:/,$p' | sed '1d')


##############################
# Run console prompts found in YAML definition.
##############################
declare -A FIELD_VALUES

if [[ -n "$FIELDS" ]]; then
  echo "Template requires additional info:"

  # Open a file descriptor to the terminal, reading and writing.
  exec 3<>/dev/tty

  while read -r LINE; do
    # Trim leading spaces
    LINE=$(echo "$LINE" | sed 's/^[[:space:]]*//')

    # Detect name. It may start with "-" dash matching yaml array syntax.
    if [[ "$LINE" =~ ^-?[[:space:]]*name:[[:space:]]*\"?([^\"]+)\"? ]]; then
      CURRENT_FIELD_NAME="${BASH_REMATCH[1]}"
      continue
    fi

    # Detect prompt. It may start with "-" dash matching yaml array syntax.
    if [[ "$LINE" =~ ^-?[[:space:]]*prompt:[[:space:]]*\"?([^\"]+)\"? ]]; then
      CURRENT_FIELD_PROMPT="${BASH_REMATCH[1]}"
    fi

    # If both name and prompt are set, ask user
    if [[ -n "$CURRENT_FIELD_NAME" && -n "$CURRENT_FIELD_PROMPT" ]]; then
      # Ask user
      printf "%s: " "$CURRENT_FIELD_PROMPT" >&3
      read -r FIELD_INPUT <&3
      FIELD_VALUES["$CURRENT_FIELD_NAME"]="$FIELD_INPUT"

      # Reset for next field
      CURRENT_FIELD_NAME=""
      CURRENT_FIELD_PROMPT=""
    fi
  done <<< "$FIELDS"

  # Close fd 3: read and write.
  exec 3<&-
  exec 3>&-
fi


##############################
# Create "Title" and "Body" from template using collected inputs.
##############################

# Build PR title with placeholders replaced
for KEY in "${!FIELD_VALUES[@]}"; do
  PR_TITLE="${PR_TITLE//\{\{$KEY\}\}/${FIELD_VALUES[$KEY]}}"
done
echo "$PR_TITLE" > "$TEMP_PR_TITLE"

# Build PR body with placeholders replaced
PR_BODY=$(sed '/^---/,/^---/d' "$TEMPLATE_FILE")
for KEY in "${!FIELD_VALUES[@]}"; do
  PR_BODY="${PR_BODY//\{\{$KEY\}\}/${FIELD_VALUES[$KEY]}}"
done
echo "$PR_BODY" > "$TEMP_PR_BODY"

echo "✔ PR title prepared: \"$PR_TITLE\""
echo "✔ PR body prepared in $TEMP_PR_BODY"

##############################
# URL encode "Title" and "Body".
##############################
# URL encode function (Bash only)
urlencode() {
  local string="$1"
  local length=${#string}
  local encoded=""

  for (( i=0; i<length; i++ )); do
    c=${string:i:1}
    case "$c" in
      [a-zA-Z0-9.~_-]) encoded+="$c" ;;
      ' ') encoded+="%20" ;;
      *) printf -v hex '%%%02X' "'$c"
         encoded+="$hex" ;;
    esac
  done
  echo "$encoded"
}

# Read PR body content
PR_BODY_RAW=$(<.git/.pr_body.tmp)
PR_BODY=$(urlencode "$PR_BODY_RAW")

# Read PR title content
PR_TITLE_RAW=$(<.git/.pr_title.tmp)
PR_TITLE=$(urlencode "$PR_TITLE_RAW")


# Ask for target branch
read -r -p "Target branch to compare against? (default: main): " TARGET_BRANCH < /dev/tty
TARGET_BRANCH=${TARGET_BRANCH:-main}

# Determine source branch and repo info
SOURCE_BRANCH=$(git rev-parse --abbrev-ref HEAD)
URL=$(git config --get remote.origin.url)
REPO_NAME=$(basename -s .git "$URL")
ORGANIZATION=$(echo "$URL" | sed -r 's/(.+):(.+)\/([^.]+)(\.git)?/\2/')

PR_URL="https://github.com/${ORGANIZATION}/${REPO_NAME}/compare/${TARGET_BRANCH}...${ORGANIZATION}:${REPO_NAME}:${SOURCE_BRANCH}?title=${PR_TITLE}&body=${PR_BODY}"

# This is useful for those working on forks:
USER_NAME=$(git config user.name)
PR_FORK_URL="https://github.com/${ORGANIZATION}/${REPO_NAME}/compare/${TARGET_BRANCH}...${USER_NAME}:${REPO_NAME}:${SOURCE_BRANCH}?title=${PR_TITLE}&body=${PR_BODY}"

echo "============================================================="
echo "Create PR:"
echo "$PR_URL"
echo "============================================================="
echo "Create PR from Fork:"
echo "$PR_FORK_URL"
echo "============================================================="
