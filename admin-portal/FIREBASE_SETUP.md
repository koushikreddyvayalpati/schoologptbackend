# Firebase Setup Guide for SchoolGPT

This guide will help you set up Firebase Authentication for your SchoolGPT admin portal.

## 1. Create a Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Create a project" or "Add project"
3. Enter project name: `schoolgpt-project` (or your preferred name)
4. Enable Google Analytics (optional)
5. Click "Create project"

## 2. Enable Authentication

1. In your Firebase project, go to **Authentication** in the left sidebar
2. Click "Get started"
3. Go to the **Sign-in method** tab
4. Enable the following providers:
   - **Email/Password**: Click to enable
   - **Google**: Click to enable and configure
     - Add your domain to authorized domains if needed
     - Download the configuration if required

## 3. Configure Firestore Database

1. Go to **Firestore Database** in the left sidebar
2. Click "Create database"
3. Choose "Start in test mode" (for development)
4. Select a location for your database
5. Click "Done"

## 4. Get Firebase Configuration

1. Go to **Project Settings** (gear icon in left sidebar)
2. Scroll down to "Your apps" section
3. Click the web icon `</>` to add a web app
4. Register your app with name: `SchoolGPT Admin Portal`
5. Copy the Firebase configuration object

## 5. Configure Environment Variables

Create a `.env.local` file in the `admin-portal` directory with your Firebase config:

```env
# Firebase Configuration
VITE_FIREBASE_API_KEY=your-api-key-here
VITE_FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your-project-id
VITE_FIREBASE_STORAGE_BUCKET=your-project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=123456789
VITE_FIREBASE_APP_ID=1:123456789:web:abcdef123456

# App Configuration
VITE_APP_NAME=SchoolGPT Admin Portal
VITE_APP_VERSION=1.0.0
```

## 6. Update Firestore Security Rules (Optional)

For production, update your Firestore security rules:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Users can read and write their own user document
    match /users/{userId} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
    
    // Admin users can read all user documents
    match /users/{userId} {
      allow read: if request.auth != null && 
        get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == 'admin';
    }
  }
}
```

## 7. Test the Setup

1. Start your development server: `npm run dev`
2. Navigate to `http://localhost:3000`
3. Try signing up with email/password
4. Try signing in with Google
5. Check the Firebase Console to see users being created

## 8. Features Included

✅ **Email/Password Authentication**
- Sign up with role selection (Admin, Teacher, Student)
- Sign in with email and password
- Password reset via email
- Email verification (automatic)

✅ **Google Authentication**
- One-click Google sign-in/sign-up
- Automatic user profile creation
- Role assignment for new Google users

✅ **User Management**
- User profiles stored in Firestore
- Role-based access control
- Profile updates and management
- Last login tracking

✅ **Security Features**
- Automatic token management
- Secure logout
- Protected routes
- Error handling

## 9. Troubleshooting

### Common Issues:

1. **"Firebase: Error (auth/unauthorized-domain)"**
   - Add your domain to authorized domains in Firebase Console
   - Go to Authentication > Settings > Authorized domains

2. **"Permission denied" in Firestore**
   - Check your Firestore security rules
   - Ensure user is authenticated

3. **Environment variables not loading**
   - Make sure `.env.local` is in the correct directory
   - Restart your development server
   - Check variable names start with `VITE_`

4. **Google Sign-in not working**
   - Verify Google provider is enabled
   - Check authorized domains
   - Ensure correct OAuth configuration

## 10. Production Deployment

Before deploying to production:

1. Update Firestore security rules for production
2. Set up proper environment variables on your hosting platform
3. Configure authorized domains for your production URL
4. Enable email verification requirements if needed
5. Set up monitoring and error tracking

## Support

If you encounter any issues, check:
- [Firebase Documentation](https://firebase.google.com/docs)
- [Firebase Console](https://console.firebase.google.com/)
- Browser developer console for error messages 