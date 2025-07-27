import { initializeApp } from 'firebase/app';
import { 
  getAuth, 
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  signOut as firebaseSignOut,
  updateProfile,
  User,
  onAuthStateChanged,
  GoogleAuthProvider,
  signInWithPopup,
  sendPasswordResetEmail,
  sendEmailVerification
} from 'firebase/auth';
import { 
  getFirestore, 
  doc, 
  setDoc, 
  getDoc, 
  updateDoc,
  serverTimestamp 
} from 'firebase/firestore';
import { getAnalytics } from 'firebase/analytics';

// Firebase configuration - schoolgpt-prod
const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY || "AIzaSyAn5ZNf6zQX8yrTTxei1LArSX3PnO9IOu4",
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN || "schoolgpt-prod.firebaseapp.com",
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID || "schoolgpt-prod",
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET || "schoolgpt-prod.firebasestorage.app",
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID || "108471903467",
  appId: import.meta.env.VITE_FIREBASE_APP_ID || "1:108471903467:web:354b5522a4e1c65c03caae",
  measurementId: import.meta.env.VITE_FIREBASE_MEASUREMENT_ID || "G-60VWMFY6X6"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);

// Initialize Firebase Authentication and get a reference to the service
export const auth = getAuth(app);

// Initialize Cloud Firestore and get a reference to the service
export const db = getFirestore(app);

// Initialize Analytics (optional)
export const analytics = getAnalytics(app);

// Google Auth Provider
export const googleProvider = new GoogleAuthProvider();
// Configure to show account selection screen
googleProvider.setCustomParameters({
  prompt: 'select_account'
});

// Add scopes for user information
googleProvider.addScope('profile');
googleProvider.addScope('email');

// Custom user interface for our app
export interface AppUser {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'teacher' | 'student';
  schoolId?: string;
  avatar?: string;
  emailVerified: boolean;
  createdAt: Date;
  lastLoginAt: Date;
}

// Auth service functions
export const authService = {
  // Sign up with email and password
  async signUp(email: string, password: string, name: string, role: 'admin' | 'teacher' | 'student', schoolId?: string): Promise<AppUser> {
    try {
      // Create user with Firebase Auth
      const userCredential = await createUserWithEmailAndPassword(auth, email, password);
      const firebaseUser = userCredential.user;

      // Update the user's display name
      await updateProfile(firebaseUser, {
        displayName: name
      });

      // Send email verification
      await sendEmailVerification(firebaseUser);

      // Create user document in Firestore
      const userData: AppUser = {
        id: firebaseUser.uid,
        email: firebaseUser.email!,
        name: name,
        role: role,
        schoolId: schoolId,
        avatar: firebaseUser.photoURL || undefined,
        emailVerified: firebaseUser.emailVerified,
        createdAt: new Date(),
        lastLoginAt: new Date()
      };

      await setDoc(doc(db, 'users', firebaseUser.uid), {
        ...userData,
        createdAt: serverTimestamp(),
        lastLoginAt: serverTimestamp()
      });

      return userData;
    } catch (error: any) {
      throw new Error(error.message || 'Failed to create account');
    }
  },

  // Sign in with email and password
  async signIn(email: string, password: string): Promise<AppUser> {
    try {
      const userCredential = await signInWithEmailAndPassword(auth, email, password);
      const firebaseUser = userCredential.user;

      // Get user data from Firestore
      const userDoc = await getDoc(doc(db, 'users', firebaseUser.uid));
      
      if (!userDoc.exists()) {
        throw new Error('User profile not found');
      }

      const userData = userDoc.data() as Omit<AppUser, 'id'>;
      
      // Update last login time
      await updateDoc(doc(db, 'users', firebaseUser.uid), {
        lastLoginAt: serverTimestamp()
      });

      return {
        id: firebaseUser.uid,
        email: firebaseUser.email!,
        name: userData.name || firebaseUser.displayName || 'User',
        role: userData.role,
        schoolId: userData.schoolId,
        avatar: firebaseUser.photoURL || userData.avatar,
        emailVerified: firebaseUser.emailVerified,
        createdAt: userData.createdAt,
        lastLoginAt: new Date()
      };
    } catch (error: any) {
      throw new Error(error.message || 'Failed to sign in');
    }
  },

  // Sign in with Google
  async signInWithGoogle(): Promise<AppUser> {
    try {
      // Try popup method first
      const result = await signInWithPopup(auth, googleProvider);
      const firebaseUser = result.user;

      // Check if user exists in Firestore
      const userDoc = await getDoc(doc(db, 'users', firebaseUser.uid));
      
      let userData: AppUser;
      
      if (userDoc.exists()) {
        // Existing user - update last login
        const existingData = userDoc.data() as Omit<AppUser, 'id'>;
        userData = {
          id: firebaseUser.uid,
          email: firebaseUser.email!,
          name: existingData.name || firebaseUser.displayName || 'User',
          role: existingData.role,
          schoolId: existingData.schoolId,
          avatar: firebaseUser.photoURL || existingData.avatar,
          emailVerified: firebaseUser.emailVerified,
          createdAt: existingData.createdAt,
          lastLoginAt: new Date()
        };
        
        await updateDoc(doc(db, 'users', firebaseUser.uid), {
          lastLoginAt: serverTimestamp(),
          emailVerified: firebaseUser.emailVerified
        });
      } else {
        // New user - create profile with default role
        userData = {
          id: firebaseUser.uid,
          email: firebaseUser.email!,
          name: firebaseUser.displayName || 'User',
          role: 'admin', // Default role for Google sign-in
          avatar: firebaseUser.photoURL || undefined,
          emailVerified: firebaseUser.emailVerified,
          createdAt: new Date(),
          lastLoginAt: new Date()
        };

        await setDoc(doc(db, 'users', firebaseUser.uid), {
          ...userData,
          createdAt: serverTimestamp(),
          lastLoginAt: serverTimestamp()
        });
      }

      return userData;
    } catch (error: any) {
      // Check if it's a COOP error and provide user guidance
      if (error.code === 'auth/popup-blocked' || 
          error.message?.includes('Cross-Origin-Opener-Policy') ||
          error.message?.includes('window.closed')) {
        throw new Error('Popup was blocked. Please allow popups for this site or try signing in with email/password instead.');
      }
      throw new Error(error.message || 'Failed to sign in with Google');
    }
  },

  // Sign out
  async signOut(): Promise<void> {
    try {
      await firebaseSignOut(auth);
    } catch (error: any) {
      throw new Error(error.message || 'Failed to sign out');
    }
  },

  // Send password reset email
  async resetPassword(email: string): Promise<void> {
    try {
      await sendPasswordResetEmail(auth, email);
    } catch (error: any) {
      throw new Error(error.message || 'Failed to send reset email');
    }
  },

  // Get current user data
  async getCurrentUser(): Promise<AppUser | null> {
    const firebaseUser = auth.currentUser;
    if (!firebaseUser) return null;

    try {
      const userDoc = await getDoc(doc(db, 'users', firebaseUser.uid));
      
      if (!userDoc.exists()) {
        return null;
      }

      const userData = userDoc.data() as Omit<AppUser, 'id'>;
      
      return {
        id: firebaseUser.uid,
        email: firebaseUser.email!,
        name: userData.name || firebaseUser.displayName || 'User',
        role: userData.role,
        schoolId: userData.schoolId,
        avatar: firebaseUser.photoURL || userData.avatar,
        emailVerified: firebaseUser.emailVerified,
        createdAt: userData.createdAt,
        lastLoginAt: userData.lastLoginAt
      };
    } catch (error) {
      console.error('Error getting current user:', error);
      return null;
    }
  },

  // Listen to auth state changes
  onAuthStateChanged: (callback: (user: AppUser | null) => void) => {
    return onAuthStateChanged(auth, async (firebaseUser) => {
      if (firebaseUser) {
        const appUser = await authService.getCurrentUser();
        callback(appUser);
      } else {
        callback(null);
      }
    });
  },

  // Update user profile
  async updateUserProfile(userId: string, updates: Partial<Pick<AppUser, 'name' | 'role' | 'schoolId' | 'avatar'>>): Promise<void> {
    try {
      await updateDoc(doc(db, 'users', userId), updates);
      
      // Update Firebase Auth profile if name is being updated
      if (updates.name && auth.currentUser) {
        await updateProfile(auth.currentUser, {
          displayName: updates.name
        });
      }
    } catch (error: any) {
      throw new Error(error.message || 'Failed to update profile');
    }
  }
};

export default authService; 