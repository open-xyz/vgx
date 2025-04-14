#!/bin/bash
# Demo script for VGX and VibePenTester integration

# Make sure environment is set up
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Created .env file from template. Edit it to add your API keys if needed."
fi

# Build the tools
echo "Building tools..."
make build

# Step 1: Initialize context with basic scan
echo -e "\n\n===== Step 1: Initialize Context ====="
./vgx --update-context=true examples/vulnerable.js

# Step 2: Scan with changes only
echo -e "\n\n===== Step 2: Scan Changed Files ====="
# For demo purposes, make a small change to vulnerable.js first
echo "// Adding a comment to trigger change detection" >> examples/vulnerable.js
./vgx --changes=true examples/vulnerable.js

# Step 3: Generate a report
echo -e "\n\n===== Step 3: Generate Security Report ====="
./vgx --report=true examples/vulnerable.js
echo "Check the reports directory for the security report"

# Step 4: Use the VibePenTester CLI
echo -e "\n\n===== Step 4: Use VibePenTester CLI ====="
./vibe scan -format text examples/vulnerable.js

# Step 5: Generate a VibePenTester report
echo -e "\n\n===== Step 5: Generate VibePenTester Report ====="
./vibe scan -output scan-results.json examples/
./vibe report -input scan-results.json -output security-report.html
echo "Security report generated: security-report.html"

echo -e "\n\nDemo completed! Review the outputs and reports." 