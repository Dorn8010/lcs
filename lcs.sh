#!/bin/bash
# lcs - library command search (Bash Version)
# Version 0.92 (Fix: MacOS Bash 3.2 Compatibility + # for comments in CSV)
# © 2025 by Alexander Dorn, MIT license

# To install:
# chmod +x lcs.sh
# sudo mv lcs.sh /usr/local/bin/

APP_VERSION="0.92"
DB_FILE="$HOME/.lcs-db.csv"
VERBOSE=0
ADD_MODE=0
REMOVE_MODE=0
EDIT_MODE=0
PRINT_MODE=0
COPY_MODE=0
FAST_CHOICE=0

# usage function
usage() {
    echo "============================================"
    echo "Library Command Search tool for CLI commands"
    echo "Store and find long commands easily"
    echo "                                Version $APP_VERSION"
    echo "© 2025 by Alexander Dorn, MIT license"
    echo "============================================"
    echo ""
    echo "Usage: lcs [options] search_term_last"
    echo ""
    echo "============================================"
    echo "Searches for a command in the description"
    echo "and offers the findings for execution"
    echo ""
    echo "The DB contains an explanation and"
    echo "the command with optional variables"
    echo ""
    echo "~/.lcs-db.csv is a ; separated CSV"
    echo "Here an example of an entry :"
    echo "Echo test;echo \"Hello\" # {\"Name\":\"World\"}"
    echo ""
    echo "Options :"
    echo "  --help, -h     Show this help"
    echo "  --version      Show version info"
    echo "  --verbose, -v  Show verbose logging"
    echo "  --fast, -f     Fast select option number (e.g. -f 2)"
    echo "  --print        Print command only"
    echo "  --copy         Copy command to clipboard"
    echo "                 (no execution)"
    echo "  --add          Add a new command"
    echo "                 Usage: lcs --add \"Desc\" \"Cmd\""
    echo "                 or interactive mode"
    echo "  --edit         Search, remove and re-add/edit"
    echo "  --remove       Search and remove a command"
    echo "  --db           Path to custom database"
    echo "                 default: ~/.lcs-db.csv"
    echo ""
    echo "Using Variables:"
    echo "  You can define variables in commands to be filled at runtime."
    echo "  Syntax: {\"Label\":\"DefaultValue\"}"
    echo "  Example: ssh -i {\"KeyFile\":\"~/.ssh/id_rsa\"} user@host"
}

# Parse arguments
SEARCH_ARGS=""
ADD_ARGS=()
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -h|--help) usage; exit 0 ;;
        --version) echo "lcs version $APP_VERSION"; exit 0 ;;
        -v|--verbose) VERBOSE=1 ;;
        -f|--fast)
            FAST_CHOICE="$2"
            shift # skip the number argument
            ;;
        --print) PRINT_MODE=1 ;;
        --copy) COPY_MODE=1 ;;
        --add) ADD_MODE=1 ;;
        --remove) REMOVE_MODE=1 ;;
        --edit) EDIT_MODE=1 ;;
        --db) DB_FILE="$2"; shift ;;
        -*) echo "Unknown option: $1"; usage; exit 1 ;;
        *) 
            if [ "$ADD_MODE" -eq 1 ]; then
                 ADD_ARGS+=("$1")
            else
                 # Bugfix: Only add space if SEARCH_ARGS is not empty
                 if [ -z "$SEARCH_ARGS" ]; then
                     SEARCH_ARGS="$1"
                 else
                     SEARCH_ARGS="$SEARCH_ARGS $1"
                 fi
            fi
            ;;
    esac
    shift
done

# Check DB exists (unless adding)
if [ ! -f "$DB_FILE" ] && [ "$ADD_MODE" -eq 0 ]; then
    echo "Database file not found: $DB_FILE"
    echo "Create it with format: Description;Command"
    exit 1
fi

if [ "$VERBOSE" -eq 1 ]; then
    echo "Using DB: $DB_FILE"
fi

# --- ADD FEATURE ---
if [ "$ADD_MODE" -eq 1 ]; then
    if [ "${#ADD_ARGS[@]}" -ge 2 ]; then
        # Use provided arguments
        NEW_DESC="${ADD_ARGS[0]}"
        # Join the rest as the command
        NEW_CMD="${ADD_ARGS[*]:1}"
    else
        # Interactive
        echo "--- Add New Command ---"
        read -p "Description: " NEW_DESC
        read -p "Command: " NEW_CMD
    fi
    
    if [[ -z "$NEW_DESC" ]] || [[ -z "$NEW_CMD" ]]; then
        echo "Error: Description and Command cannot be empty."
        exit 1
    fi
    
    # Append to CSV
    echo "${NEW_DESC};${NEW_CMD}" >> "$DB_FILE"
    echo "Entry added successfully."
    exit 0
fi

# Search (grep)
# We use grep with line numbers (-n) to help with removal/edit - lines starting with # count as comment only
MATCHES=$(grep -n -i "$SEARCH_ARGS" "$DB_FILE" | grep -v "^[0-9]*:#")

if [ -z "$MATCHES" ]; then
    echo "No matches found."
    exit 0
fi

# Setup array for selection
declare -a COMMANDS
declare -a DESCRIPTIONS
declare -a LINE_NUMBERS
COUNT=0

# Read grep output
while IFS= read -r line; do
    COUNT=$((COUNT+1))
    
    # Extract Line Number
    L_NUM=$(echo "$line" | cut -d: -f1)
    
    # Extract Content
    CONTENT=$(echo "$line" | cut -d: -f2-)
    DESC=$(echo "$CONTENT" | awk -F';' '{print $1}')
    CMD=$(echo "$CONTENT" | awk -F';' '{print $2}')
    
    COMMANDS[$COUNT]="$CMD"
    DESCRIPTIONS[$COUNT]="$DESC"
    LINE_NUMBERS[$COUNT]="$L_NUM"
done <<< "$MATCHES"

# --- SELECTION LOGIC ---
if [ "$FAST_CHOICE" -gt 0 ]; then
    # --- FAST SELECTION MODE ---
    if [ "$FAST_CHOICE" -gt "$COUNT" ]; then
        echo "Error: Fast choice $FAST_CHOICE is out of range. Only $COUNT matches found."
        exit 1
    fi
    SELECTION=$FAST_CHOICE
    if [ "$VERBOSE" -eq 1 ]; then
         echo "Fast selected [$SELECTION]: ${DESCRIPTIONS[$SELECTION]}"
    fi

elif [ "$COUNT" -eq 1 ]; then
    # AUTO SELECT
    SELECTION=1
    echo "Found 1 match: ${DESCRIPTIONS[1]}"
    if [ "$VERBOSE" -eq 1 ]; then
         echo "Cmd: ${COMMANDS[1]}"
    fi
else
    # SHOW MENU
    if [ "$REMOVE_MODE" -eq 1 ]; then
        echo "Select command to REMOVE:"
    elif [ "$EDIT_MODE" -eq 1 ]; then
        echo "Select command to EDIT:"
    else
        echo "Found commands:"
    fi

    for (( i=1; i<=COUNT; i++ ))
    do
        echo "[$i] ${DESCRIPTIONS[$i]}"
        echo "    ${COMMANDS[$i]}"
    done

    echo ""
    read -p "Select a number: " SELECTION

    if [[ ! "$SELECTION" =~ ^[0-9]+$ ]] || [ "$SELECTION" -lt 1 ] || [ "$SELECTION" -gt "$COUNT" ]; then
        echo "Invalid selection."
        exit 1
    fi
fi

# --- REMOVE / EDIT FEATURE ---
if [ "$REMOVE_MODE" -eq 1 ] || [ "$EDIT_MODE" -eq 1 ]; then
    TARGET_LINE=${LINE_NUMBERS[$SELECTION]}
    OLD_DESC="${DESCRIPTIONS[$SELECTION]}"
    OLD_CMD="${COMMANDS[$SELECTION]}"
    
    # Remove the line (safely with tmp file)
    sed "${TARGET_LINE}d" "$DB_FILE" > "${DB_FILE}.tmp" && mv "${DB_FILE}.tmp" "$DB_FILE"
    
    if [ "$REMOVE_MODE" -eq 1 ]; then
        echo "Entry removed successfully."
        exit 0
    fi
    
    # EDIT MODE Logic
    if [ "$EDIT_MODE" -eq 1 ]; then
        echo "--- Edit Entry (Press Enter to keep current) ---"
        
        # MacOS / Bash 3.2 Compatible logic (No -i flag)
        echo "Current Desc: $OLD_DESC"
        read -p "New Desc: " NEW_DESC
        if [ -z "$NEW_DESC" ]; then
            NEW_DESC="$OLD_DESC"
        fi

        echo "Current Cmd: $OLD_CMD"
        read -p "New Cmd: " NEW_CMD
        if [ -z "$NEW_CMD" ]; then
            NEW_CMD="$OLD_CMD"
        fi
        
        echo "${NEW_DESC};${NEW_CMD}" >> "$DB_FILE"
        echo "Entry updated successfully."
        exit 0
    fi
fi

# Normal Execution Flow
FINAL_CMD="${COMMANDS[$SELECTION]}"

# --- Parameter Parsing Logic ---
# Defined outside loop for compatibility
VAR_REGEX='\{("[^"]+"):[[:space:]]*("[^"]*")\}'

while [[ "$FINAL_CMD" =~ $VAR_REGEX ]]; do
    FULL_MATCH="${BASH_REMATCH[0]}" 
    RAW_KEY="${BASH_REMATCH[1]}" 
    RAW_VAL="${BASH_REMATCH[2]}" 

    # --- FIX FOR BASH 3.2 (MacOS) ---
    # Bash 3.2 does not support negative substring indexing like ${VAR:1:-1}
    # We must calculate length manually.
    
    # Calculate length - 2 (to remove surrounding quotes)
    KLEN=$((${#RAW_KEY}-2))
    VLEN=$((${#RAW_VAL}-2))
    
    # Extract Substring
    KEY="${RAW_KEY:1:$KLEN}"
    VAL="${RAW_VAL:1:$VLEN}"

    # MacOS / Bash 3.2 Compatible logic
    read -e -p "Input for '$KEY' [$VAL]: " USER_INPUT
    
    if [ -z "$USER_INPUT" ]; then
        USER_INPUT="$VAL"
    fi
    
    # ESCAPING FOR SED:
    # 1. Escape the PATTERN (FULL_MATCH). 
    SAFE_MATCH=$(printf '%s\n' "$FULL_MATCH" | sed 's/[][\.*^$/]/\\&/g')

    # 2. Escape the REPLACEMENT (USER_INPUT). 
    #    We must escape only: \ / & (do NOT escape . or *)
    #    Note: We escape backslash first!
    SAFE_INPUT=$(printf '%s\n' "$USER_INPUT" | sed 's/\\/\\\\/g; s/\//\\\//g; s/&/\\&/g')
    
    # Replace ONLY the first occurrence
    FINAL_CMD=$(echo "$FINAL_CMD" | sed "s/$SAFE_MATCH/$SAFE_INPUT/")
done

# --- FEATURE: PRINT ONLY ---
if [ "$PRINT_MODE" -eq 1 ]; then
    echo "$FINAL_CMD"
    exit 0
fi

# --- FEATURE: COPY TO CLIPBOARD ---
if [ "$COPY_MODE" -eq 1 ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        echo -n "$FINAL_CMD" | pbcopy
        echo "Command copied to clipboard."
    elif command -v xclip &> /dev/null; then
        # Linux xclip
        echo -n "$FINAL_CMD" | xclip -selection clipboard
        echo "Command copied to clipboard."
    elif command -v xsel &> /dev/null; then
        # Linux xsel
        echo -n "$FINAL_CMD" | xsel --clipboard --input
        echo "Command copied to clipboard."
    else
        echo "Error: Neither 'pbcopy', 'xclip' nor 'xsel' found."
        echo "Command was:"
        echo "$FINAL_CMD"
        exit 1
    fi
    exit 0
fi

# --- EXECUTION ---
if [ "$VERBOSE" -eq 1 ]; then
    echo "Executing: $FINAL_CMD"
else
    if [ "$COUNT" -gt 1 ]; then
        echo "Executing..."
    fi
fi

# Using exec replaces the script process with the command.
# This solves the 'signal: killed' issue with SSH/interactive tools.
exec bash -c "$FINAL_CMD"
