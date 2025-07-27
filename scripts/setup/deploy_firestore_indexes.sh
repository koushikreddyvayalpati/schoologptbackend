#!/bin/bash

# Deploy Firestore Indexes Script
# This script deploys the composite indexes defined in firestore.indexes.json

set -e

echo "🔥 Deploying Firestore Indexes for SchoolGPT..."

# Check if Firebase CLI is installed
if ! command -v firebase &> /dev/null; then
    echo "❌ Firebase CLI is not installed. Please install it first:"
    echo "npm install -g firebase-tools"
    exit 1
fi

# Check if logged in to Firebase
if ! firebase projects:list &> /dev/null; then
    echo "❌ Not logged in to Firebase. Please login first:"
    echo "firebase login"
    exit 1
fi

# Check if firestore.indexes.json exists
if [ ! -f "firestore.indexes.json" ]; then
    echo "❌ firestore.indexes.json not found in current directory"
    echo "Please run this script from the project root directory"
    exit 1
fi

# Get project ID (you may need to set this)
PROJECT_ID=${FIREBASE_PROJECT_ID:-"schoolgpt-demo"}

echo "📋 Project ID: $PROJECT_ID"
echo "📁 Indexes file: firestore.indexes.json"

# Validate the indexes file
echo "🔍 Validating indexes configuration..."
if ! python3 -m json.tool firestore.indexes.json > /dev/null 2>&1; then
    echo "❌ Invalid JSON in firestore.indexes.json"
    exit 1
fi

echo "✅ Indexes configuration is valid"

# Deploy the indexes
echo "🚀 Deploying Firestore indexes..."
echo "This may take several minutes as Firestore builds the indexes..."

firebase deploy --only firestore:indexes --project $PROJECT_ID

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Firestore indexes deployed successfully!"
    echo ""
    echo "📊 Index Status:"
    echo "You can check the status of your indexes at:"
    echo "https://console.firebase.google.com/project/$PROJECT_ID/firestore/indexes"
    echo ""
    echo "⏱️  Note: Indexes may take several minutes to build depending on your data size."
    echo "Complex queries will be enabled once the indexes are in 'Enabled' state."
    echo ""
    echo "🎯 Optimized Queries Available:"
    echo "- Student attendance patterns (student_id + date)"
    echo "- Teacher class analytics (teacher_id + subject)"
    echo "- Date range attendance queries (date + marked_by)"
    echo "- Multi-field student searches (parent_id + grade)"
    echo "- Assignment and grade analytics"
    echo ""
else
    echo "❌ Failed to deploy Firestore indexes"
    exit 1
fi 