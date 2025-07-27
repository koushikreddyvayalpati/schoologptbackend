import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { 
  Users, 
  GraduationCap, 
  TrendingUp, 
  Calendar,
  Plus,
  ArrowUpRight,
  Zap,
  BookOpen,
  BarChart3,
  Clock,
  CheckCircle,
  ArrowRight,
  UserPlus,
  ClipboardList,
  Award,
  Target,
  Bell,
  AlertTriangle,
  Activity,
  FileText,
  MessageSquare
} from 'lucide-react'
import { useAppSelector } from '../store/hooks'

// School Admin Data
const adminStats = [
  {
    name: 'Total Students',
    value: '847',
    change: '+12',
    changeValue: 'this month',
    trend: 'up',
    icon: GraduationCap,
    color: 'text-blue-600',
    bgGradient: 'from-blue-500 to-blue-600',
    description: 'Enrolled students',
  },
  {
    name: 'Teaching Staff',
    value: '42',
    change: '+3',
    changeValue: 'new hires',
    trend: 'up',
    icon: Users,
    color: 'text-green-600',
    bgGradient: 'from-green-500 to-emerald-600',
    description: 'Active teachers',
  },
  {
    name: 'Attendance Rate',
    value: '94.2%',
    change: '+2.1%',
    changeValue: 'vs last month',
    trend: 'up',
    icon: CheckCircle,
    color: 'text-purple-600',
    bgGradient: 'from-purple-500 to-purple-600',
    description: 'Daily average',
  },
  {
    name: 'Academic Performance',
    value: '87.5%',
    change: '+4.3%',
    changeValue: 'improvement',
    trend: 'up',
    icon: TrendingUp,
    color: 'text-orange-600',
    bgGradient: 'from-orange-500 to-red-500',
    description: 'Overall grades',
  },
]

// Teacher Data
const teacherStats = [
  {
    name: 'My Classes',
    value: '6',
    change: '156',
    changeValue: 'total students',
    trend: 'up',
    icon: Users,
    color: 'text-blue-600',
    bgGradient: 'from-blue-500 to-blue-600',
    description: 'Active classes',
  },
  {
    name: 'Pending Assignments',
    value: '23',
    change: '12',
    changeValue: 'due today',
    trend: 'up',
    icon: ClipboardList,
    color: 'text-orange-600',
    bgGradient: 'from-orange-500 to-red-500',
    description: 'To review',
  },
  {
    name: 'Attendance Today',
    value: '92%',
    change: '144/156',
    changeValue: 'present',
    trend: 'up',
    icon: CheckCircle,
    color: 'text-green-600',
    bgGradient: 'from-green-500 to-emerald-600',
    description: 'Students present',
  },
  {
    name: 'Class Average',
    value: '85.2%',
    change: '+3.1%',
    changeValue: 'this quarter',
    trend: 'up',
    icon: Award,
    color: 'text-purple-600',
    bgGradient: 'from-purple-500 to-purple-600',
    description: 'Grade average',
  },
]

// Student Data
const studentStats = [
  {
    name: 'Current GPA',
    value: '3.8',
    change: '+0.2',
    changeValue: 'this semester',
    trend: 'up',
    icon: Award,
    color: 'text-blue-600',
    bgGradient: 'from-blue-500 to-blue-600',
    description: 'Grade point average',
  },
  {
    name: 'Assignments Due',
    value: '4',
    change: '2',
    changeValue: 'due tomorrow',
    trend: 'up',
    icon: ClipboardList,
    color: 'text-orange-600',
    bgGradient: 'from-orange-500 to-red-500',
    description: 'Pending work',
  },
  {
    name: 'Attendance Rate',
    value: '96%',
    change: '23/24',
    changeValue: 'days this month',
    trend: 'up',
    icon: CheckCircle,
    color: 'text-green-600',
    bgGradient: 'from-green-500 to-emerald-600',
    description: 'Days present',
  },
  {
    name: 'Learning Progress',
    value: '78%',
    change: '+12%',
    changeValue: 'this quarter',
    trend: 'up',
    icon: Target,
    color: 'text-purple-600',
    bgGradient: 'from-purple-500 to-purple-600',
    description: 'Goals completed',
  },
]

// Quick Actions by Role
const adminQuickActions = [
  {
    name: 'Create User Accounts',
    description: 'Add new students, teachers, or staff',
    href: '/create-accounts',
    icon: UserPlus,
    color: 'primary',
    bgGradient: 'from-blue-500 to-purple-600',
    featured: true,
  },
  {
    name: 'View Reports',
    description: 'Academic and administrative analytics',
    href: '/analytics',
    icon: BarChart3,
    color: 'secondary',
    bgGradient: 'from-green-500 to-emerald-600',
    featured: false,
  },
  {
    name: 'School Calendar',
    description: 'Manage events and schedules',
    href: '/calendar',
    icon: Calendar,
    color: 'success',
    bgGradient: 'from-orange-500 to-red-500',
    featured: false,
  },
]

const teacherQuickActions = [
  {
    name: 'Take Attendance',
    description: 'Mark student attendance for today',
    href: '/attendance',
    icon: CheckCircle,
    color: 'primary',
    bgGradient: 'from-green-500 to-emerald-600',
    featured: true,
  },
  {
    name: 'Add Students',
    description: 'Enroll students to your classes',
    href: '/add-students',
    icon: UserPlus,
    color: 'secondary',
    bgGradient: 'from-blue-500 to-purple-600',
    featured: true,
  },
  {
    name: 'Create Assignment',
    description: 'New homework or project',
    href: '/assignments/create',
    icon: Plus,
    color: 'success',
    bgGradient: 'from-orange-500 to-red-500',
    featured: false,
  },
]

const studentQuickActions = [
  {
    name: 'View Assignments',
    description: 'Check homework and due dates',
    href: '/assignments',
    icon: ClipboardList,
    color: 'primary',
    bgGradient: 'from-blue-500 to-purple-600',
    featured: true,
  },
  {
    name: 'Check Progress',
    description: 'View learning analytics',
    href: '/progress',
    icon: TrendingUp,
    color: 'secondary',
    bgGradient: 'from-green-500 to-emerald-600',
    featured: true,
  },
  {
    name: 'Study Goals',
    description: 'Set and track learning targets',
    href: '/goals',
    icon: Target,
    color: 'success',
    bgGradient: 'from-purple-500 to-pink-600',
    featured: false,
  },
]

const Dashboard: React.FC = () => {
  const { user } = useAppSelector((state) => state.auth)

  // Get role-specific data
  const getStatsData = () => {
    switch (user?.role) {
      case 'admin':
        return adminStats
      case 'teacher':
        return teacherStats
      case 'student':
        return studentStats
      default:
        return adminStats
    }
  }

  const getQuickActions = () => {
    switch (user?.role) {
      case 'admin':
        return adminQuickActions
      case 'teacher':
        return teacherQuickActions
      case 'student':
        return studentQuickActions
      default:
        return adminQuickActions
    }
  }

  const getGreeting = () => {
    switch (user?.role) {
      case 'admin':
        return 'Good morning, Administrator! 👨‍💼'
      case 'teacher':
        return 'Good morning, Teacher! 👩‍🏫'
      case 'student':
        return 'Good morning, Student! 👨‍🎓'
      default:
        return 'Good morning! 👋'
    }
  }

  const getDescription = () => {
    switch (user?.role) {
      case 'admin':
        return "Here's your school overview for today."
      case 'teacher':
        return "Ready for another great day of teaching?"
      case 'student':
        return "Let's make today a productive learning day!"
      default:
        return "Welcome to your dashboard."
    }
  }

  const stats = getStatsData()
  const quickActions = getQuickActions()

  // Recent Activity (role-specific)
  const getRecentActivity = () => {
    switch (user?.role) {
      case 'admin':
        return [
          {
            id: 1,
            action: 'New teacher account created',
            details: 'Sarah Johnson - Mathematics',
            time: '2 hours ago',
            type: 'success',
            icon: UserPlus,
            avatar: '👩‍🏫',
          },
          {
            id: 2,
            action: 'Student enrollment completed',
            details: '15 new students added to Grade 9',
            time: '4 hours ago',
            type: 'info',
            icon: GraduationCap,
            avatar: '📚',
          },
          {
            id: 3,
            action: 'Monthly report generated',
            details: 'Academic performance analytics',
            time: '6 hours ago',
            type: 'success',
            icon: BarChart3,
            avatar: '📊',
          },
        ]
      case 'teacher':
        return [
          {
            id: 1,
            action: 'Assignment submitted',
            details: 'Math Quiz - Grade 10A (23/25 students)',
            time: '1 hour ago',
            type: 'success',
            icon: ClipboardList,
            avatar: '✅',
          },
          {
            id: 2,
            action: 'New student added',
            details: 'Emma Wilson joined Grade 10B',
            time: '3 hours ago',
            type: 'info',
            icon: UserPlus,
            avatar: '👧',
          },
          {
            id: 3,
            action: 'Lesson plan updated',
            details: 'Algebra - Chapter 5 modifications',
            time: '5 hours ago',
            type: 'info',
            icon: BookOpen,
            avatar: '📖',
          },
        ]
      case 'student':
        return [
          {
            id: 1,
            action: 'Assignment graded',
            details: 'History Essay - A- (87%)',
            time: '2 hours ago',
            type: 'success',
            icon: Award,
            avatar: '🎯',
          },
          {
            id: 2,
            action: 'New assignment posted',
            details: 'Science Lab Report due Friday',
            time: '4 hours ago',
            type: 'warning',
            icon: ClipboardList,
            avatar: '🧪',
          },
          {
            id: 3,
            action: 'Goal achieved',
            details: 'Reading comprehension target reached',
            time: '1 day ago',
            type: 'success',
            icon: Target,
            avatar: '🎉',
          },
        ]
      default:
        return []
    }
  }

  const recentActivity = getRecentActivity()

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Enhanced page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900">
            {getGreeting()}
          </h1>
          <p className="text-gray-600 text-lg">
            {getDescription()}
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 text-sm text-gray-500 bg-white px-4 py-2 rounded-xl shadow-sm border border-gray-100">
            <Calendar className="w-4 h-4" />
            <span>{new Date().toLocaleDateString('en-US', { 
              weekday: 'long', 
              year: 'numeric', 
              month: 'long', 
              day: 'numeric' 
            })}</span>
          </div>
          {user?.role === 'admin' && (
            <button className="btn-primary">
              <UserPlus className="w-4 h-4 mr-2" />
              Add User
            </button>
          )}
          {user?.role === 'teacher' && (
            <button className="btn-primary">
              <CheckCircle className="w-4 h-4 mr-2" />
              Take Attendance
            </button>
          )}
          {user?.role === 'student' && (
            <button className="btn-primary">
              <Target className="w-4 h-4 mr-2" />
              Set Goal
            </button>
          )}
        </div>
      </div>

      {/* Enhanced stats grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat, index) => {
          const Icon = stat.icon
          return (
            <div
              key={stat.name}
              className="card-hover group animate-slide-up"
              style={{ animationDelay: `${index * 0.1}s` }}
            >
              <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${stat.bgGradient} flex items-center justify-center shadow-lg`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <div className="text-right">
                    <div className="flex items-center text-sm font-medium text-green-600">
                      <ArrowUpRight className="w-4 h-4 mr-1" />
                      {stat.change}
                    </div>
                    <div className="text-xs text-gray-500">{stat.changeValue}</div>
                  </div>
                </div>
                <div className="space-y-1">
                  <h3 className="text-2xl font-bold text-gray-900 group-hover:text-blue-600 transition-colors">
                    {stat.value}
                  </h3>
                  <p className="text-sm font-medium text-gray-600">{stat.name}</p>
                  <p className="text-xs text-gray-500">{stat.description}</p>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Enhanced quick actions */}
      <div className="card animate-slide-up" style={{ animationDelay: '0.4s' }}>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900">
              Quick Actions
            </h2>
            <Zap className="w-5 h-5 text-yellow-500" />
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            {quickActions.map((action, index) => {
              const Icon = action.icon
              return (
                <a
                  key={action.name}
                  href={action.href}
                  className={`group relative overflow-hidden rounded-2xl p-6 transition-all duration-300 hover:scale-105 ${
                    action.featured 
                      ? 'bg-gradient-to-br from-blue-50 to-purple-50 border-2 border-blue-200 shadow-lg shadow-blue-500/10' 
                      : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'
                  }`}
                >
                  {action.featured && (
                    <div className="absolute top-2 right-2">
                      <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
                    </div>
                  )}
                  <div className="flex items-start space-x-4">
                    <div className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${action.bgGradient} flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform`}>
                      <Icon className="w-6 h-6 text-white" />
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
                        {action.name}
                      </h3>
                      <p className="text-sm text-gray-600 mt-1">
                        {action.description}
                      </p>
                      <div className="flex items-center mt-3 text-sm text-blue-600 group-hover:text-blue-700">
                        <span>Get started</span>
                        <ArrowRight className="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform" />
                      </div>
                    </div>
                  </div>
                </a>
              )
            })}
          </div>
        </div>
      </div>

      {/* Two column layout */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Enhanced recent activity */}
        <div className="lg:col-span-2 card animate-slide-up" style={{ animationDelay: '0.6s' }}>
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">
                Recent Activity
              </h2>
              <button className="btn-ghost text-sm">
                View all
                <ArrowRight className="w-4 h-4 ml-1" />
              </button>
            </div>
            <div className="space-y-4">
              {recentActivity.map((activity, index) => {
                const Icon = activity.icon
                return (
                  <div
                    key={activity.id}
                    className="flex items-start space-x-4 p-4 rounded-xl hover:bg-gray-50 transition-colors group"
                  >
                    <div className="flex-shrink-0">
                      <div className="w-10 h-10 bg-gray-100 rounded-xl flex items-center justify-center text-lg">
                        {activity.avatar}
                      </div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-gray-900 group-hover:text-blue-600 transition-colors">
                          {activity.action}
                        </p>
                        <div className="flex items-center text-xs text-gray-500">
                          <Clock className="w-3 h-3 mr-1" />
                          {activity.time}
                        </div>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">{activity.details}</p>
                    </div>
                    <div className="flex-shrink-0">
                      <Icon className={`w-5 h-5 ${
                        activity.type === 'success' ? 'text-green-500' :
                        activity.type === 'warning' ? 'text-yellow-500' :
                        activity.type === 'info' ? 'text-blue-500' : 'text-gray-400'
                      }`} />
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {/* Role-specific sidebar */}
        <div className="card animate-slide-up" style={{ animationDelay: '0.8s' }}>
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">
                {user?.role === 'admin' ? 'School Overview' : 
                 user?.role === 'teacher' ? 'Today\'s Classes' : 
                 'My Schedule'}
              </h2>
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
            </div>
            <div className="space-y-4">
              {user?.role === 'admin' && (
                <>
                  <div className="p-4 rounded-xl border border-gray-200 hover:border-blue-200 hover:bg-blue-50/50 transition-all">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-sm font-medium text-gray-900">Total Classes</h3>
                        <p className="text-2xl font-bold text-blue-600">48</p>
                      </div>
                      <BookOpen className="w-8 h-8 text-blue-500" />
                    </div>
                  </div>
                  <div className="p-4 rounded-xl border border-gray-200 hover:border-green-200 hover:bg-green-50/50 transition-all">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-sm font-medium text-gray-900">Active Courses</h3>
                        <p className="text-2xl font-bold text-green-600">24</p>
                      </div>
                      <GraduationCap className="w-8 h-8 text-green-500" />
                    </div>
                  </div>
                  <div className="p-4 rounded-xl border border-gray-200 hover:border-purple-200 hover:bg-purple-50/50 transition-all">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-sm font-medium text-gray-900">Departments</h3>
                        <p className="text-2xl font-bold text-purple-600">8</p>
                      </div>
                      <Users className="w-8 h-8 text-purple-500" />
                    </div>
                  </div>
                </>
              )}
              
              {user?.role === 'teacher' && (
                <>
                  <div className="p-4 rounded-xl bg-blue-50 border border-blue-200">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-sm font-medium text-blue-900">Next Class</h3>
                        <p className="text-lg font-bold text-blue-700">Math - Grade 10A</p>
                        <p className="text-sm text-blue-600">Room 201 • 10:30 AM</p>
                      </div>
                      <Clock className="w-8 h-8 text-blue-500" />
                    </div>
                  </div>
                  <div className="p-4 rounded-xl border border-gray-200">
                    <h3 className="text-sm font-medium text-gray-900 mb-2">Today's Schedule</h3>
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span>Grade 9B - Algebra</span>
                        <span className="text-gray-500">9:00 AM</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span>Grade 10A - Geometry</span>
                        <span className="text-gray-500">10:30 AM</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span>Grade 11C - Calculus</span>
                        <span className="text-gray-500">2:00 PM</span>
                      </div>
                    </div>
                  </div>
                </>
              )}
              
              {user?.role === 'student' && (
                <>
                  <div className="p-4 rounded-xl bg-green-50 border border-green-200">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-sm font-medium text-green-900">Next Class</h3>
                        <p className="text-lg font-bold text-green-700">Chemistry</p>
                        <p className="text-sm text-green-600">Lab 3 • 11:00 AM</p>
                      </div>
                      <Clock className="w-8 h-8 text-green-500" />
                    </div>
                  </div>
                  <div className="p-4 rounded-xl border border-gray-200">
                    <h3 className="text-sm font-medium text-gray-900 mb-2">Upcoming Deadlines</h3>
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span>History Essay</span>
                        <span className="text-red-500">Tomorrow</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span>Math Problem Set</span>
                        <span className="text-yellow-500">Friday</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span>Science Project</span>
                        <span className="text-green-500">Next Week</span>
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard 