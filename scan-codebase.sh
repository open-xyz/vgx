#!/bin/bash
set -e

# Help message
function show_help {
    echo "VGX Codebase Scanner"
    echo "---------------------"
    echo "Usage: $0 [options] [path_to_codebase]"
    echo ""
    echo "Options:"
    echo "  -a, --all             Scan all files, not just changed ones"
    echo "  -f, --file FILE       Scan a specific file or directory"
    echo "  -v, --vibe            Enable VibePenTester integration"
    echo "  -r, --report FORMAT   Generate report in specified format (html, md)"
    echo "  -o, --openai KEY      Specify OpenAI API key"
    echo "  -k, --vibe-key KEY    Specify VibePenTester API key"
    echo "  -s, --server URL      Specify VibePenTester server URL"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 /path/to/your/project"
    echo "  $0 --all --vibe /path/to/your/project"
    echo "  $0 --file src/main.js /path/to/your/project"
    echo ""
    exit 0
}

# Default values
SCAN_TARGET="$(pwd)"
SCAN_CMD="scan-changes"
VIBE_ENABLED="false"
REPORT_FORMAT=""
SPECIFIC_FILE=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--all)
            SCAN_CMD="scan"
            shift
            ;;
        -f|--file)
            SPECIFIC_FILE="$2"
            shift 2
            ;;
        -v|--vibe)
            VIBE_ENABLED="true"
            shift
            ;;
        -r|--report)
            REPORT_FORMAT="$2"
            shift 2
            ;;
        -o|--openai)
            export OPENAI_API_KEY="$2"
            shift 2
            ;;
        -k|--vibe-key)
            export VIBE_API_KEY="$2"
            shift 2
            ;;
        -s|--server)
            export VIBE_SERVER_URL="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            ;;
        *)
            # If it's the last argument and not an option, treat it as the target directory
            if [[ $# -eq 1 && ! $1 =~ ^- ]]; then
                SCAN_TARGET="$1"
            fi
            shift
            ;;
    esac
done

# Ensure the target directory exists
if [ ! -d "$SCAN_TARGET" ]; then
    echo "Error: Target directory '$SCAN_TARGET' does not exist."
    exit 1
fi

# Adjust command if a specific file is provided
if [ -n "$SPECIFIC_FILE" ]; then
    SCAN_CMD="scan $SPECIFIC_FILE"
fi

# Ensure Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "Error: Docker is not running or not accessible."
    exit 1
fi

# Build and run the scanner
echo "Building VGX scanner container..."
export SCAN_TARGET
export SCAN_CMD
export VIBE_ENABLED

# Check if the image already exists, if not build it
if ! docker images | grep -q "vgx-scanner"; then
    docker-compose build
fi

echo "Starting scan of '$SCAN_TARGET'..."
docker-compose up --remove-orphans

# If report format is specified, generate the appropriate report
if [ -n "$REPORT_FORMAT" ]; then
    echo "Generating $REPORT_FORMAT report..."
    
    # Find the latest JSON report
    LATEST_JSON=$(find "$SCAN_TARGET/reports" -name "security-report-*.json" -type f -printf "%T@ %p\n" | sort -n | tail -1 | cut -d' ' -f2-)
    
    if [ -z "$LATEST_JSON" ]; then
        echo "Error: No JSON report found to convert."
        exit 1
    fi
    
    OUTPUT_FILE="$SCAN_TARGET/reports/security-report-$(date +%Y%m%d-%H%M%S).$REPORT_FORMAT"
    
    docker-compose run --rm vgx-scanner vibe report --input "$LATEST_JSON" --format "$REPORT_FORMAT" --output "$OUTPUT_FILE"
    
    echo "Report generated: $OUTPUT_FILE"
fi

echo "Scan complete! Check the 'reports' directory in your codebase for findings." 