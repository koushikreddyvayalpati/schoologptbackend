import React, { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from './store/hooks'
import { initializeAuth } from './store/slices/authSlice'
import Layout from './components/Layout'
import AuthLoader from './components/ui/AuthLoader'

// Auth pages
import SignIn from './pages/SignIn'
import SignUp from './pages/SignUp'
import ForgotPassword from './pages/ForgotPassword'

// Common pages
import Dashboard from './pages/Dashboard'
import Settings from './pages/Settings'

// Admin pages
import Schools from './pages/Schools'
import SchoolSetup from './pages/SchoolSetup'
import CreateAccounts from './pages/CreateAccounts'

// Teacher pages
import TeacherClasses from './pages/TeacherClasses'

// Student pages
import StudentProgress from './pages/StudentProgress'

// Protected Route Component
const ProtectedRoute: React.FC<{ children: React.ReactNode; allowedRoles?: string[] }> = ({ 
  children, 
  allowedRoles 
}) => {
  const { isAuthenticated, user } = useAppSelector((state) => state.auth)
  
  if (!isAuthenticated) {
    return <Navigate to="/signin" replace />
  }
  
  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return <Navigate to="/dashboard" replace />
  }
  
  return <>{children}</>
}

// Public Route Component (redirect if authenticated)
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAppSelector((state) => state.auth)
  
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }
  
  return <>{children}</>
}

// Assignment Page Component (handles role-based content)
const AssignmentPage: React.FC = () => {
  const { user } = useAppSelector((state) => state.auth)
  
  return (
    <div className="p-8 text-center">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Assignments</h1>
      <p className="text-gray-600">
        {user?.role === 'teacher' 
          ? 'Create and manage assignments'
          : 'View and submit your assignments'
        }
      </p>
    </div>
  )
}

// Messages Page Component (handles role-based content)
const MessagesPage: React.FC = () => {
  const { user } = useAppSelector((state) => state.auth)
  
  return (
    <div className="p-8 text-center">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Messages</h1>
      <p className="text-gray-600">
        {user?.role === 'teacher' 
          ? 'Communicate with students and parents'
          : 'Chat with your teachers'
        }
      </p>
    </div>
  )
}

function App() {
  const dispatch = useAppDispatch()
  const { isLoading } = useAppSelector((state) => state.auth)

  useEffect(() => {
    dispatch(initializeAuth())
  }, [dispatch])

  if (isLoading) {
    return <AuthLoader />
  }

  return (
    <Routes>
      {/* Public Routes */}
      <Route path="/signin" element={
        <PublicRoute>
          <SignIn />
        </PublicRoute>
      } />
      <Route path="/signup" element={
        <PublicRoute>
          <SignUp />
        </PublicRoute>
      } />
      <Route path="/forgot-password" element={
        <PublicRoute>
          <ForgotPassword />
        </PublicRoute>
      } />
      
      {/* Protected Routes */}
      <Route path="/" element={
        <ProtectedRoute>
          <Layout />
        </ProtectedRoute>
      }>
        {/* Common Routes */}
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="settings" element={<Settings />} />
        
        {/* Admin Only Routes */}
        <Route path="schools" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <Schools />
          </ProtectedRoute>
        } />
        <Route path="school-setup" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <SchoolSetup />
          </ProtectedRoute>
        } />
        <Route path="create-accounts" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <CreateAccounts />
          </ProtectedRoute>
        } />
        <Route path="staff" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Staff Management</h1>
              <p className="text-gray-600">Manage your teaching staff and administrators</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="students" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Student Management</h1>
              <p className="text-gray-600">Manage student records and enrollment</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="curriculum" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Curriculum Management</h1>
              <p className="text-gray-600">Manage courses and academic programs</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="analytics" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">School Analytics</h1>
              <p className="text-gray-600">View comprehensive school performance data</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="calendar" element={
          <ProtectedRoute allowedRoles={['admin', 'teacher']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">School Calendar</h1>
              <p className="text-gray-600">Manage school events and schedules</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="communications" element={
          <ProtectedRoute allowedRoles={['admin']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Communications</h1>
              <p className="text-gray-600">Send announcements and manage communications</p>
            </div>
          </ProtectedRoute>
        } />
        
        {/* Teacher Only Routes */}
        <Route path="classes" element={
          <ProtectedRoute allowedRoles={['teacher']}>
            <TeacherClasses />
          </ProtectedRoute>
        } />
        <Route path="add-students" element={
          <ProtectedRoute allowedRoles={['teacher']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Add Students</h1>
              <p className="text-gray-600">Enroll students to your classes</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="attendance" element={
          <ProtectedRoute allowedRoles={['teacher']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Attendance</h1>
              <p className="text-gray-600">Take and manage student attendance</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="assignments" element={
          <ProtectedRoute allowedRoles={['teacher', 'student']}>
            <AssignmentPage />
          </ProtectedRoute>
        } />
        <Route path="gradebook" element={
          <ProtectedRoute allowedRoles={['teacher']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Gradebook</h1>
              <p className="text-gray-600">Grade student work and track progress</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="lesson-plans" element={
          <ProtectedRoute allowedRoles={['teacher']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Lesson Plans</h1>
              <p className="text-gray-600">Create and organize your lesson plans</p>
            </div>
          </ProtectedRoute>
        } />
        
        {/* Student Only Routes */}
        <Route path="courses" element={
          <ProtectedRoute allowedRoles={['student']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">My Courses</h1>
              <p className="text-gray-600">View your enrolled courses and materials</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="grades" element={
          <ProtectedRoute allowedRoles={['student']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">My Grades</h1>
              <p className="text-gray-600">View your academic performance</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="progress" element={
          <ProtectedRoute allowedRoles={['student']}>
            <StudentProgress />
          </ProtectedRoute>
        } />
        <Route path="schedule" element={
          <ProtectedRoute allowedRoles={['student']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">My Schedule</h1>
              <p className="text-gray-600">View your class timetable</p>
            </div>
          </ProtectedRoute>
        } />
        <Route path="goals" element={
          <ProtectedRoute allowedRoles={['student']}>
            <div className="p-8 text-center">
              <h1 className="text-2xl font-bold text-gray-900 mb-4">Learning Goals</h1>
              <p className="text-gray-600">Set and track your learning objectives</p>
            </div>
          </ProtectedRoute>
        } />
        
        {/* Shared Routes */}
        <Route path="messages" element={
          <ProtectedRoute allowedRoles={['teacher', 'student']}>
            <MessagesPage />
          </ProtectedRoute>
        } />
        
        {/* Catch all route */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  )
}

export default App 