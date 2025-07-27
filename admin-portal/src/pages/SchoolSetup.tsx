import React, { useState, useEffect } from 'react'
import { useAppSelector } from '../store/hooks'
import { 
  Sparkles, 
  ArrowRight, 
  CheckCircle, 
  School,
  Users,
  BookOpen,
  MessageSquare,
  Zap,
  Bot,
  Settings,
  Database
} from 'lucide-react'
import AISchoolSetupChat from '../components/AISchoolSetupChat'
import apiService from '../services/api'

interface SetupStep {
  id: string
  title: string
  description: string
  icon: React.ComponentType<any>
  completed: boolean
}

const SchoolSetup: React.FC = () => {
  const { user } = useAppSelector((state) => state.auth)
  const [currentStep, setCurrentStep] = useState<'welcome' | 'chat' | 'configure' | 'complete'>('welcome')
  const [setupData, setSetupData] = useState<any>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [apiHealth, setApiHealth] = useState<boolean | null>(null)

  // Check backend connection on mount
  useEffect(() => {
    checkBackendConnection()
  }, [])

  const checkBackendConnection = async () => {
    try {
      const response = await apiService.healthCheck()
      setApiHealth(response.success)
      setIsConnected(response.success)
    } catch (error) {
      console.error('Backend connection failed:', error)
      setApiHealth(false)
      setIsConnected(false)
    }
  }

  const steps: SetupStep[] = [
    {
      id: 'chat',
      title: 'AI-Powered Setup',
      description: 'Chat with our AI assistant to configure your school',
      icon: Bot,
      completed: currentStep === 'complete' || (currentStep !== 'welcome' && currentStep !== 'chat')
    },
    {
      id: 'configure',
      title: 'System Configuration',
      description: 'Review and finalize your school settings',
      icon: Settings,
      completed: currentStep === 'complete'
    },
    {
      id: 'database',
      title: 'Database Setup',
      description: 'Initialize your school database',
      icon: Database,
      completed: currentStep === 'complete'
    },
    {
      id: 'users',
      title: 'Create Accounts',
      description: 'Add teachers and students to your system',
      icon: Users,
      completed: currentStep === 'complete'
    }
  ]

  const handleChatComplete = (chatData: any) => {
    setSetupData(chatData)
    setCurrentStep('configure')
  }

  const handleConfigurationComplete = async () => {
    try {
      // Save setup data to backend
      const response = await apiService.startSchoolSetup({
        schoolName: setupData?.schoolName || 'My School',
        adminName: user?.name || 'Admin',
        adminEmail: user?.email || 'admin@school.edu',
        description: 'School setup via AI assistant',
      })

      if (response.success) {
        setCurrentStep('complete')
      }
    } catch (error) {
      console.error('Failed to complete setup:', error)
    }
  }

  const renderConnectionStatus = () => (
    <div className="mb-6">
      <div className={`p-4 rounded-xl border-2 ${
        isConnected 
          ? 'bg-green-50 border-green-200' 
          : 'bg-yellow-50 border-yellow-200'
      }`}>
        <div className="flex items-center space-x-3">
          <div className={`w-3 h-3 rounded-full ${
            isConnected ? 'bg-green-500' : 'bg-yellow-500'
          } animate-pulse`}></div>
          <div>
            <p className={`font-medium ${
              isConnected ? 'text-green-800' : 'text-yellow-800'
            }`}>
              Backend Status: {isConnected ? 'Connected' : 'Connecting...'}
            </p>
            <p className={`text-sm ${
              isConnected ? 'text-green-600' : 'text-yellow-600'
            }`}>
              {isConnected 
                ? 'AI features and data sync are available'
                : 'Some features may be limited'
              }
            </p>
          </div>
        </div>
      </div>
    </div>
  )

  const renderWelcome = () => (
    <div className="max-w-4xl mx-auto">
      {renderConnectionStatus()}
      
      <div className="text-center mb-12">
        <div className="w-20 h-20 bg-gradient-to-br from-blue-600 to-purple-600 rounded-3xl flex items-center justify-center mx-auto mb-6 shadow-2xl">
          <Sparkles className="w-10 h-10 text-white" />
        </div>
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Welcome to SchoolGPT Setup! 🎓
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          Let's set up your school with the power of AI. Our intelligent assistant will guide you through the process.
        </p>
        <p className="text-lg text-gray-500">
          Hello <span className="font-semibold text-blue-600">{user?.name}</span>! Ready to create an amazing educational experience?
        </p>
      </div>

      {/* Setup Steps Preview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
        {steps.map((step, index) => {
          const Icon = step.icon
          return (
            <div key={step.id} className="relative">
              <div className="card p-6 text-center hover:shadow-lg transition-all">
                <div className={`w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4 ${
                  step.completed 
                    ? 'bg-green-100 text-green-600' 
                    : 'bg-blue-100 text-blue-600'
                }`}>
                  {step.completed ? (
                    <CheckCircle className="w-8 h-8" />
                  ) : (
                    <Icon className="w-8 h-8" />
                  )}
                </div>
                <h3 className="font-semibold text-gray-900 mb-2">{step.title}</h3>
                <p className="text-sm text-gray-600">{step.description}</p>
              </div>
              
              {index < steps.length - 1 && (
                <div className="hidden lg:block absolute top-1/2 -right-3 transform -translate-y-1/2">
                  <ArrowRight className="w-6 h-6 text-gray-300" />
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Features Preview */}
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-3xl p-8 mb-8 border border-blue-200">
        <h2 className="text-2xl font-bold text-gray-900 mb-6 text-center">
          What You'll Get
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="text-center">
            <div className="w-12 h-12 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <MessageSquare className="w-6 h-6 text-blue-600" />
            </div>
            <h3 className="font-semibold text-gray-900 mb-2">AI Assistant</h3>
            <p className="text-gray-600 text-sm">Intelligent setup guidance and ongoing support</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-green-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <Users className="w-6 h-6 text-green-600" />
            </div>
            <h3 className="font-semibold text-gray-900 mb-2">User Management</h3>
            <p className="text-gray-600 text-sm">Easy teacher and student account creation</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-purple-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <BookOpen className="w-6 h-6 text-purple-600" />
            </div>
            <h3 className="font-semibold text-gray-900 mb-2">Educational Tools</h3>
            <p className="text-gray-600 text-sm">Complete classroom management system</p>
          </div>
        </div>
      </div>

      {/* Start Button */}
      <div className="text-center">
        <button
          onClick={() => setCurrentStep('chat')}
          className="btn-primary text-lg px-8 py-4 shadow-xl hover:shadow-2xl transform hover:scale-105 transition-all"
        >
          <Zap className="w-6 h-6 mr-3" />
          Start AI Setup
        </button>
        <p className="text-sm text-gray-500 mt-4">
          Setup typically takes 5-10 minutes
        </p>
      </div>
    </div>
  )

  const renderChat = () => (
    <div className="max-w-4xl mx-auto">
      {renderConnectionStatus()}
      
      <div className="mb-8">
        <div className="flex items-center space-x-4 mb-4">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-600 to-purple-600 rounded-2xl flex items-center justify-center">
            <Bot className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">AI School Setup</h1>
            <p className="text-gray-600">Chat with our AI to configure your school</p>
          </div>
        </div>
        
        {/* Progress Steps */}
        <div className="flex items-center space-x-2 mb-6">
          {steps.map((step, index) => (
            <div key={step.id} className="flex items-center">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                step.completed 
                  ? 'bg-green-500 text-white' 
                  : index === 0 ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-600'
              }`}>
                {step.completed ? '✓' : index + 1}
              </div>
              {index < steps.length - 1 && (
                <div className={`w-8 h-1 mx-2 ${step.completed ? 'bg-green-500' : 'bg-gray-200'}`}></div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* AI Chat Component */}
      <div className="h-[600px]">
        <AISchoolSetupChat
          adminName={user?.name || 'Admin'}
          adminEmail={user?.email || 'admin@school.edu'}
          onSetupComplete={handleChatComplete}
        />
      </div>

      {/* Help Section */}
      <div className="mt-8 p-6 bg-gray-50 rounded-2xl">
        <h3 className="font-semibold text-gray-900 mb-3">Having trouble?</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-gray-600 mb-2">💡 <strong>Tips:</strong></p>
            <ul className="text-gray-600 space-y-1">
              <li>• Be specific about your school details</li>
              <li>• Use the suggested responses for faster setup</li>
              <li>• Ask questions if you need clarification</li>
            </ul>
          </div>
          <div>
            <p className="text-gray-600 mb-2">🔧 <strong>Technical:</strong></p>
            <ul className="text-gray-600 space-y-1">
              <li>• Backend API: {isConnected ? '✅ Connected' : '❌ Disconnected'}</li>
              <li>• AI Features: {isConnected ? '✅ Available' : '⚠️ Limited'}</li>
              <li>• Data Sync: {isConnected ? '✅ Real-time' : '❌ Offline mode'}</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )

  const renderConfigure = () => (
    <div className="max-w-4xl mx-auto">
      {renderConnectionStatus()}
      
      <div className="text-center mb-8">
        <div className="w-16 h-16 bg-gradient-to-br from-green-500 to-emerald-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <Settings className="w-8 h-8 text-white" />
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          Review & Configure
        </h1>
        <p className="text-gray-600">
          Review your AI-generated setup and make any final adjustments
        </p>
      </div>

      <div className="card p-8 mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-6">Setup Summary</h2>
        <div className="space-y-4">
          <div className="p-4 bg-blue-50 rounded-xl">
            <h3 className="font-medium text-blue-900 mb-2">Chat Summary</h3>
            <p className="text-blue-700 text-sm">
              {setupData?.messages?.length ? 
                `Completed ${setupData.messages.length} interactions with AI assistant` :
                'AI setup conversation completed successfully'
              }
            </p>
          </div>
          <div className="p-4 bg-green-50 rounded-xl">
            <h3 className="font-medium text-green-900 mb-2">School Information</h3>
            <p className="text-green-700 text-sm">
              Admin: {user?.name} ({user?.email})
            </p>
          </div>
        </div>
      </div>

      <div className="flex justify-between">
        <button
          onClick={() => setCurrentStep('chat')}
          className="btn-secondary"
        >
          Back to Chat
        </button>
        <button
          onClick={handleConfigurationComplete}
          className="btn-primary"
        >
          Complete Setup
          <ArrowRight className="w-4 h-4 ml-2" />
        </button>
      </div>
    </div>
  )

  const renderComplete = () => (
    <div className="max-w-4xl mx-auto text-center">
      <div className="w-24 h-24 bg-gradient-to-br from-green-500 to-emerald-600 rounded-3xl flex items-center justify-center mx-auto mb-8 shadow-2xl">
        <CheckCircle className="w-12 h-12 text-white" />
      </div>
      
      <h1 className="text-4xl font-bold text-gray-900 mb-4">
        🎉 Setup Complete!
      </h1>
      <p className="text-xl text-gray-600 mb-8">
        Your school has been successfully configured with SchoolGPT
      </p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
        <div className="card p-6 text-center">
          <div className="w-12 h-12 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <Users className="w-6 h-6 text-blue-600" />
          </div>
          <h3 className="font-semibold text-gray-900 mb-2">Create Accounts</h3>
          <p className="text-gray-600 text-sm mb-4">Add teachers and students to your school</p>
          <button className="btn-secondary text-sm">
            Add Users
          </button>
        </div>
        
        <div className="card p-6 text-center">
          <div className="w-12 h-12 bg-green-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <BookOpen className="w-6 h-6 text-green-600" />
          </div>
          <h3 className="font-semibold text-gray-900 mb-2">Setup Curriculum</h3>
          <p className="text-gray-600 text-sm mb-4">Configure subjects and courses</p>
          <button className="btn-secondary text-sm">
            Manage Curriculum
          </button>
        </div>
        
        <div className="card p-6 text-center">
          <div className="w-12 h-12 bg-purple-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <MessageSquare className="w-6 h-6 text-purple-600" />
          </div>
          <h3 className="font-semibold text-gray-900 mb-2">AI Assistant</h3>
          <p className="text-gray-600 text-sm mb-4">Get ongoing help and support</p>
          <button className="btn-secondary text-sm">
            Chat with AI
          </button>
        </div>
      </div>

      <button
        onClick={() => window.location.href = '/dashboard'}
        className="btn-primary text-lg px-8 py-4"
      >
        Go to Dashboard
        <ArrowRight className="w-5 h-5 ml-2" />
      </button>
    </div>
  )

  const renderCurrentStep = () => {
    switch (currentStep) {
      case 'welcome':
        return renderWelcome()
      case 'chat':
        return renderChat()
      case 'configure':
        return renderConfigure()
      case 'complete':
        return renderComplete()
      default:
        return renderWelcome()
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
      {renderCurrentStep()}
    </div>
  )
}

export default SchoolSetup 