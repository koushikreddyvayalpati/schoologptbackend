import React from 'react'
import { NavLink } from 'react-router-dom'
import { 
  Home, 
  Users, 
  UserPlus, 
  Settings, 
  BarChart3, 
  X,
  Sparkles,
  BookOpen,
  Calendar,
  MessageSquare,
  Shield,
  GraduationCap,
  ClipboardList,
  Award,
  Bell,
  FileText,
  Clock,
  Target,
  TrendingUp,
  CheckSquare,
  Book,
  PlusCircle,
  UserCheck,
  LogOut,
  ChevronRight,
  Building2
} from 'lucide-react'
import { useAppSelector } from '../store/hooks'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

// School Admin Navigation
const adminNavigation = [
  { 
    name: 'Dashboard', 
    href: '/dashboard', 
    icon: Home,
    description: 'School overview',
    color: 'text-blue-600',
    bgColor: 'bg-blue-50'
  },
  { 
    name: 'Staff Management', 
    href: '/staff', 
    icon: Users,
    description: 'Teachers & staff',
    color: 'text-purple-600',
    bgColor: 'bg-purple-50'
  },
  { 
    name: 'Student Management', 
    href: '/students', 
    icon: GraduationCap,
    description: 'Student records',
    color: 'text-green-600',
    bgColor: 'bg-green-50'
  },
  { 
    name: 'Create Accounts', 
    href: '/create-accounts', 
    icon: UserPlus,
    description: 'Add users',
    color: 'text-orange-600',
    bgColor: 'bg-orange-50',
    featured: true
  },
  { 
    name: 'Schools', 
    href: '/schools', 
    icon: Building2,
    description: 'Manage schools',
    color: 'text-blue-600',
    bgColor: 'bg-blue-50'
  },
  { 
    name: 'School Setup', 
    href: '/school-setup', 
    icon: Sparkles,
    description: 'AI-powered setup',
    color: 'text-emerald-600',
    bgColor: 'bg-emerald-50',
    featured: true
  },
  { 
    name: 'Curriculum', 
    href: '/curriculum', 
    icon: BookOpen,
    description: 'Course management',
    color: 'text-teal-600',
    bgColor: 'bg-teal-50'
  },
  { 
    name: 'Analytics', 
    href: '/analytics', 
    icon: BarChart3,
    description: 'School performance',
    color: 'text-indigo-600',
    bgColor: 'bg-indigo-50'
  },
  { 
    name: 'Calendar', 
    href: '/calendar', 
    icon: Calendar,
    description: 'School events',
    color: 'text-red-600',
    bgColor: 'bg-red-50'
  },
  { 
    name: 'Communications', 
    href: '/communications', 
    icon: MessageSquare,
    description: 'Announcements',
    color: 'text-pink-600',
    bgColor: 'bg-pink-50'
  },
]

// Teacher Navigation
const teacherNavigation = [
  { 
    name: 'Dashboard', 
    href: '/dashboard', 
    icon: Home,
    description: 'Teaching overview',
    color: 'text-blue-600',
    bgColor: 'bg-blue-50'
  },
  { 
    name: 'My Classes', 
    href: '/classes', 
    icon: Users,
    description: 'Manage classes',
    color: 'text-purple-600',
    bgColor: 'bg-purple-50'
  },
  { 
    name: 'Add Students', 
    href: '/add-students', 
    icon: UserCheck,
    description: 'Enroll students',
    color: 'text-green-600',
    bgColor: 'bg-green-50',
    featured: true
  },
  { 
    name: 'Attendance', 
    href: '/attendance', 
    icon: CheckSquare,
    description: 'Daily attendance',
    color: 'text-orange-600',
    bgColor: 'bg-orange-50'
  },
  { 
    name: 'Assignments', 
    href: '/assignments', 
    icon: ClipboardList,
    description: 'Create & grade',
    color: 'text-teal-600',
    bgColor: 'bg-teal-50'
  },
  { 
    name: 'Gradebook', 
    href: '/gradebook', 
    icon: Award,
    description: 'Student grades',
    color: 'text-indigo-600',
    bgColor: 'bg-indigo-50'
  },
  { 
    name: 'Lesson Plans', 
    href: '/lesson-plans', 
    icon: Book,
    description: 'Plan lessons',
    color: 'text-red-600',
    bgColor: 'bg-red-50'
  },
  { 
    name: 'Messages', 
    href: '/messages', 
    icon: MessageSquare,
    description: 'Student & parent chat',
    color: 'text-pink-600',
    bgColor: 'bg-pink-50'
  },
]

// Student Navigation
const studentNavigation = [
  { 
    name: 'Dashboard', 
    href: '/dashboard', 
    icon: Home,
    description: 'My overview',
    color: 'text-blue-600',
    bgColor: 'bg-blue-50'
  },
  { 
    name: 'My Courses', 
    href: '/courses', 
    icon: BookOpen,
    description: 'Current classes',
    color: 'text-purple-600',
    bgColor: 'bg-purple-50'
  },
  { 
    name: 'Assignments', 
    href: '/assignments', 
    icon: ClipboardList,
    description: 'Homework & tasks',
    color: 'text-green-600',
    bgColor: 'bg-green-50'
  },
  { 
    name: 'Grades', 
    href: '/grades', 
    icon: Award,
    description: 'My performance',
    color: 'text-orange-600',
    bgColor: 'bg-orange-50'
  },
  { 
    name: 'Progress', 
    href: '/progress', 
    icon: TrendingUp,
    description: 'Learning analytics',
    color: 'text-teal-600',
    bgColor: 'bg-teal-50',
    featured: true
  },
  { 
    name: 'Schedule', 
    href: '/schedule', 
    icon: Calendar,
    description: 'Class timetable',
    color: 'text-indigo-600',
    bgColor: 'bg-indigo-50'
  },
  { 
    name: 'Goals', 
    href: '/goals', 
    icon: Target,
    description: 'Learning targets',
    color: 'text-red-600',
    bgColor: 'bg-red-50'
  },
  { 
    name: 'Messages', 
    href: '/messages', 
    icon: MessageSquare,
    description: 'Teacher chat',
    color: 'text-pink-600',
    bgColor: 'bg-pink-50'
  },
]

const bottomNavigation = [
  { 
    name: 'Settings', 
    href: '/settings', 
    icon: Settings,
    description: 'Preferences',
    color: 'text-gray-600',
    bgColor: 'bg-gray-50'
  },
]

const Sidebar: React.FC<SidebarProps> = ({ isOpen, onClose }) => {
  const { user } = useAppSelector((state) => state.auth)
  
  // Get navigation based on user role
  const getNavigation = () => {
    switch (user?.role) {
      case 'admin':
        return adminNavigation
      case 'teacher':
        return teacherNavigation
      case 'student':
        return studentNavigation
      default:
        return adminNavigation
    }
  }

  const navigation = getNavigation()

  const getRoleDisplayName = () => {
    switch (user?.role) {
      case 'admin':
        return 'School Admin'
      case 'teacher':
        return 'Teacher'
      case 'student':
        return 'Student'
      default:
        return 'User'
    }
  }

  const getRoleIcon = () => {
    switch (user?.role) {
      case 'admin':
        return Shield
      case 'teacher':
        return BookOpen
      case 'student':
        return GraduationCap
      default:
        return Shield
    }
  }

  const RoleIcon = getRoleIcon()

  return (
    <>
      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-72 lg:flex-col">
        <div className="flex min-h-0 flex-1 flex-col glass border-r border-gray-200/50 backdrop-blur-xl">
          {/* Logo */}
          <div className="flex h-16 flex-shrink-0 items-center px-6 border-b border-gray-200/50">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
                <Sparkles className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gradient-primary">
                  SchoolGPT
                </h1>
                <p className="text-xs text-gray-500">AI-Powered Education</p>
              </div>
            </div>
          </div>

          {/* Role indicator */}
          <div className="px-6 py-4 border-b border-gray-200/50">
            <div className="flex items-center space-x-3">
              <div className={`w-8 h-8 rounded-xl flex items-center justify-center ${
                user?.role === 'admin' ? 'bg-blue-100 text-blue-600' :
                user?.role === 'teacher' ? 'bg-green-100 text-green-600' :
                'bg-purple-100 text-purple-600'
              }`}>
                <RoleIcon className="w-4 h-4" />
              </div>
              <div>
                <p className="text-sm font-semibold text-gray-900">{getRoleDisplayName()}</p>
                <p className="text-xs text-gray-500">{user?.name || 'User'}</p>
              </div>
            </div>
          </div>

          {/* Navigation */}
          <div className="flex flex-1 flex-col overflow-y-auto pt-6 pb-4">
            <nav className="flex-1 space-y-2 px-4">
              {navigation.map((item) => (
                <NavLink
                  key={item.name}
                  to={item.href}
                  className={({ isActive }) =>
                    `nav-link group relative ${
                      isActive
                        ? 'nav-link-active'
                        : 'nav-link-inactive'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      {item.featured && !isActive && (
                        <div className="absolute -top-1 -right-1">
                          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                        </div>
                      )}
                      <div className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all ${
                        isActive 
                          ? `${item.bgColor} ${item.color}` 
                          : 'bg-gray-100 text-gray-400 group-hover:bg-gray-200 group-hover:text-gray-600'
                      }`}>
                        <item.icon className="w-5 h-5" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-medium truncate">
                            {item.name}
                          </span>
                          {item.featured && (
                            <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                              NEW
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 truncate mt-0.5">
                          {item.description}
                        </p>
                      </div>
                    </>
                  )}
                </NavLink>
              ))}
            </nav>

            {/* Bottom navigation */}
            <div className="mt-8 px-4">
              <div className="border-t border-gray-200/50 pt-4">
                {bottomNavigation.map((item) => (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    className={({ isActive }) =>
                      `nav-link group ${
                        isActive
                          ? 'nav-link-active'
                          : 'nav-link-inactive'
                      }`
                    }
                  >
                    {({ isActive }) => (
                      <>
                        <div className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all ${
                          isActive 
                            ? `${item.bgColor} ${item.color}` 
                            : 'bg-gray-100 text-gray-400 group-hover:bg-gray-200 group-hover:text-gray-600'
                        }`}>
                          <item.icon className="w-5 h-5" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <span className="text-sm font-medium truncate">
                            {item.name}
                          </span>
                          <p className="text-xs text-gray-500 truncate mt-0.5">
                            {item.description}
                          </p>
                        </div>
                      </>
                    )}
                  </NavLink>
                ))}
              </div>
            </div>

            {/* School info section */}
            <div className="px-4 py-4">
              <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-2xl p-4 border border-blue-200/50">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl flex items-center justify-center">
                    <GraduationCap className="w-4 h-4 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900">{user?.schoolId || 'My School'}</p>
                    <p className="text-xs text-gray-500">Academic Year 2024</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile sidebar - similar structure */}
      <div
        className={`fixed inset-y-0 left-0 z-50 w-72 glass backdrop-blur-xl transform transition-transform duration-300 ease-in-out lg:hidden ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex min-h-0 flex-1 flex-col border-r border-gray-200/50">
          {/* Mobile logo with close button */}
          <div className="flex h-16 flex-shrink-0 items-center justify-between px-6 border-b border-gray-200/50">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
                <Sparkles className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gradient-primary">
                  SchoolGPT
                </h1>
                <p className="text-xs text-gray-500">AI-Powered Education</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-xl text-gray-400 hover:text-gray-500 hover:bg-gray-100 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Mobile role indicator */}
          <div className="px-6 py-4 border-b border-gray-200/50">
            <div className="flex items-center space-x-3">
              <div className={`w-8 h-8 rounded-xl flex items-center justify-center ${
                user?.role === 'admin' ? 'bg-blue-100 text-blue-600' :
                user?.role === 'teacher' ? 'bg-green-100 text-green-600' :
                'bg-purple-100 text-purple-600'
              }`}>
                <RoleIcon className="w-4 h-4" />
              </div>
              <div>
                <p className="text-sm font-semibold text-gray-900">{getRoleDisplayName()}</p>
                <p className="text-xs text-gray-500">{user?.name || 'User'}</p>
              </div>
            </div>
          </div>

          {/* Mobile navigation */}
          <div className="flex flex-1 flex-col overflow-y-auto pt-6 pb-4">
            <nav className="flex-1 space-y-2 px-4">
              {navigation.map((item) => (
                <NavLink
                  key={item.name}
                  to={item.href}
                  onClick={onClose}
                  className={({ isActive }) =>
                    `nav-link group relative ${
                      isActive
                        ? 'nav-link-active'
                        : 'nav-link-inactive'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      {item.featured && !isActive && (
                        <div className="absolute -top-1 -right-1">
                          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                        </div>
                      )}
                      <div className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all ${
                        isActive 
                          ? `${item.bgColor} ${item.color}` 
                          : 'bg-gray-100 text-gray-400 group-hover:bg-gray-200 group-hover:text-gray-600'
                      }`}>
                        <item.icon className="w-5 h-5" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-medium truncate">
                            {item.name}
                          </span>
                          {item.featured && (
                            <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                              NEW
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 truncate mt-0.5">
                          {item.description}
                        </p>
                      </div>
                    </>
                  )}
                </NavLink>
              ))}
            </nav>

            {/* Mobile bottom navigation */}
            <div className="mt-8 px-4">
              <div className="border-t border-gray-200/50 pt-4">
                {bottomNavigation.map((item) => (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    onClick={onClose}
                    className={({ isActive }) =>
                      `nav-link group ${
                        isActive
                          ? 'nav-link-active'
                          : 'nav-link-inactive'
                      }`
                    }
                  >
                    {({ isActive }) => (
                      <>
                        <div className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all ${
                          isActive 
                            ? `${item.bgColor} ${item.color}` 
                            : 'bg-gray-100 text-gray-400 group-hover:bg-gray-200 group-hover:text-gray-600'
                        }`}>
                          <item.icon className="w-5 h-5" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <span className="text-sm font-medium truncate">
                            {item.name}
                          </span>
                          <p className="text-xs text-gray-500 truncate mt-0.5">
                            {item.description}
                          </p>
                        </div>
                      </>
                    )}
                  </NavLink>
                ))}
              </div>
            </div>

            {/* Mobile school info section */}
            <div className="px-4 py-4">
              <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-2xl p-4 border border-blue-200/50">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl flex items-center justify-center">
                    <GraduationCap className="w-4 h-4 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900">{user?.schoolId || 'My School'}</p>
                    <p className="text-xs text-gray-500">Academic Year 2024</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default Sidebar 