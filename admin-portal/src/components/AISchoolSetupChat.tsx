import React, { useState, useRef, useEffect } from 'react';
import {
  Bot,
  Send,
  Loader2,
  MessageCircle,
  CheckCircle2,
  User,
  Sparkles,
  School,
  AlertCircle,
  Zap,
  Upload,
  FileSpreadsheet,
  Download,
  Settings,
  Users,
  BookOpen,
  Calendar,
  BarChart3,
  Building2
} from 'lucide-react';
import apiService, { ChatResponse } from '../services/api';

interface Message {
  id: string;
  type: 'user' | 'ai' | 'system';
  content: string;
  timestamp: Date;
  suggestions?: string[];
  isLoading?: boolean;
  attachments?: {
    type: 'file' | 'template' | 'config';
    name: string;
    description?: string;
    downloadUrl?: string;
  }[];
}

interface AISchoolSetupChatProps {
  onSetupComplete?: (setupData: any) => void;
  adminName: string;
  adminEmail: string;
}

const AISchoolSetupChat: React.FC<AISchoolSetupChatProps> = ({
  onSetupComplete,
  adminName,
  adminEmail
}) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const [setupComplete, setSetupComplete] = useState(false);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [showFileUpload, setShowFileUpload] = useState(false);
  const [uploadedFiles, setUploadedFiles] = useState<File[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Initialize chat when component mounts
  useEffect(() => {
    initializeChat();
  }, []);

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const initializeChat = async () => {
    try {
      setIsLoading(true);
      
      // Start a new chat session with the setup agent
      const response = await apiService.startSetupChat({ 
        admin_name: adminName, 
        admin_email: adminEmail 
      });
      
      if (response.success && response.data) {
        const sessionData = response.data;
        
        if (sessionData.session_id) {
          setSessionId(sessionData.session_id);
        }
        
        // Enhanced welcome message with more capabilities
        const welcomeMessage: Message = {
          id: '1',
          type: 'ai',
          content: sessionData.welcome_message || `👋 Hello ${adminName}! I'm your AI School Setup Assistant.

I can help you:
🏫 **Quick Setup**: Tell me about your school and I'll configure everything
📊 **Bulk Import**: Upload Excel/CSV files with student, teacher, or parent data
⚙️ **Custom Configuration**: Fine-tune settings for your specific needs
🎯 **Smart Suggestions**: Get recommendations based on your school type

Let's start! What's your school's name and what type of institution is it?`,
          timestamp: new Date(),
          suggestions: [
            "🏫 My school is Greenwood International School with 1200 students in India",
            "📚 We're setting up Springfield Elementary for grades K-5",
            "🎓 Create Oak Valley High School with CBSE curriculum",
            "📊 I have Excel files with student and teacher data to upload",
            "⚙️ Show me advanced configuration options"
          ]
        };
        
        setMessages([welcomeMessage]);
        setIsInitialized(true);
      } else {
        throw new Error(response.error || 'Failed to start chat session');
      }
    } catch (error) {
      // Show connection error instead of fake AI response
      const errorMessage: Message = {
        id: '1',
        type: 'system',
        content: `❌ **Connection Error**: Unable to start AI assistant session.

The school setup requires connection to our AI backend service. Please:
• Check your internet connection
• Verify the backend service is running
• Contact support if the issue persists

**Note**: All intelligent responses should come from the Gemini-powered backend, not the frontend.`,
        timestamp: new Date(),
        suggestions: [
          "🔄 Retry connection",
          "🔧 Check backend service status",
          "📞 Contact technical support"
        ]
      };
      
      setMessages([errorMessage]);
      setIsInitialized(true);
    } finally {
      setIsLoading(false);
    }
  };



  const handleFileUpload = () => {
    setShowFileUpload(true);
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files) {
      const newFiles = Array.from(files);
      setUploadedFiles(prev => [...prev, ...newFiles]);
      
      // Add system message about uploaded files
      const fileMessage: Message = {
        id: Date.now().toString(),
        type: 'system',
        content: `📁 Uploaded ${newFiles.length} file(s): ${newFiles.map(f => f.name).join(', ')}`,
        timestamp: new Date(),
        attachments: newFiles.map(file => ({
          type: 'file' as const,
          name: file.name,
          description: `${file.size > 1024 * 1024 ? (file.size / (1024 * 1024)).toFixed(1) + 'MB' : (file.size / 1024).toFixed(1) + 'KB'}`
        }))
      };
      
      setMessages(prev => [...prev, fileMessage]);
      
      // Auto-send message about files
      sendMessage(`I've uploaded ${newFiles.length} file(s) with data: ${newFiles.map(f => f.name).join(', ')}. Please process this data for school setup.`);
    }
    setShowFileUpload(false);
  };

  const downloadTemplate = (templateType: string) => {
    // In a real implementation, this would download actual templates from backend
    const templates = {
      'students': 'student_import_template.xlsx',
      'teachers': 'teacher_import_template.xlsx', 
      'parents': 'parent_import_template.xlsx',
      'classes': 'class_structure_template.xlsx'
    };
    
    // Show system message about template download (not fake AI response)
    const templateMessage: Message = {
      id: Date.now().toString(),
      type: 'system',
      content: `📥 **Template Download**: ${templateType} import template ready for download.

Fill in your data using the template format and upload it for bulk import.`,
      timestamp: new Date(),
      attachments: [{
        type: 'template' as const,
        name: templates[templateType as keyof typeof templates] || 'template.xlsx',
        description: `Template for importing ${templateType} data`,
        downloadUrl: `#download-${templateType}-template`
      }]
    };
    
    setMessages(prev => [...prev, templateMessage]);
    
    // TODO: Actual template download should call backend API
    // apiService.downloadTemplate(templateType)
  };

  const sendMessage = async (content: string) => {
    if (!content.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: content.trim(),
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);

    const loadingMessage: Message = {
      id: (Date.now() + 1).toString(),
      type: 'ai',
      content: '',
      timestamp: new Date(),
      isLoading: true,
    };

    setMessages(prev => [...prev, loadingMessage]);

    try {
      if (!sessionId) {
        throw new Error('No active session. Please refresh the page.');
      }

      let response;
      
      // Enhanced confirmation detection
      const isConfirmation = content.toLowerCase().includes('yes') || 
                           content.toLowerCase().includes('confirm') || 
                           content.toLowerCase().includes('create the school') ||
                           content.toLowerCase().includes('proceed') ||
                           content.toLowerCase().includes('continue with');

      const lastAiMessage = messages.slice().reverse().find(m => m.type === 'ai');
      const isConfirmationPending = lastAiMessage?.content.includes('Does this look correct?') || 
                                   lastAiMessage?.content.includes('Reply with:') ||
                                   lastAiMessage?.content.includes('This configuration will be created');

      // Handle continuation commands
      const isContinuation = content.toLowerCase().includes('continue') ||
                           content.toLowerCase().includes('proceed') ||
                           content.toLowerCase().includes('go ahead') ||
                           content.toLowerCase().includes('that looks good');

      if ((isConfirmation || isContinuation) && isConfirmationPending) {
        response = await apiService.confirmConfiguration(sessionId, content);
      } else {
        response = await apiService.sendChatMessage(sessionId, content);
      }

      if (response.success && response.data) {
        const responseData = response.data;
        
        const aiResponse: Message = {
          id: (Date.now() + 2).toString(),
          type: 'ai',
          content: responseData.message || responseData.agent_response || responseData.response || "I understand. Could you tell me more about your school?",
          timestamp: new Date(),
          suggestions: responseData.suggestions || [] // Only use suggestions from backend AI
        };

        setMessages(prev => prev.slice(0, -1).concat(aiResponse));

        if (responseData.status === 'completed' || responseData.session_status === 'completed') {
          setSetupComplete(true);
          
          // Add completion message with next steps but KEEP the chat open
          const completionMessage: Message = {
            id: (Date.now() + 3).toString(),
            type: 'ai',
            content: `🎉 **Setup Complete!** 

${responseData.school_name || 'Your school'} has been successfully created!

**School Details:**
• 🏫 School ID: ${responseData.school_id || 'Generated'}
• 🔑 School Code: ${responseData.school_code || 'Generated'}
• 📍 Region: ${responseData.region || 'Configured'}
• 🎓 Education System: ${responseData.education_system || 'CBSE'}

**What's Next?**
The school is ready! You can continue chatting with me to:`,
            timestamp: new Date(),
            suggestions: [
              "📊 Upload student data via Excel/CSV files",
              "👨‍🏫 Create teacher accounts in bulk",
              "👨‍👩‍👧‍👦 Import parent/guardian information",
              "⚙️ Configure additional school settings",
              "📚 Set up class schedules and subjects",
              "🎯 Go to dashboard and start using the system"
            ]
          };
          
          setMessages(prev => [...prev.slice(0, -1), aiResponse, completionMessage]);
          
          // DON'T call onSetupComplete to avoid closing the chat
          // Keep the chat open for further interactions
          return;
        } else if (responseData.confirmation_pending) {
          // For confirmation pending, use backend suggestions or basic confirmation options
          const confirmMessage = {
            ...aiResponse,
            suggestions: responseData.suggestions || [
              "✅ Yes, create the school with these settings", 
              "📝 Let me modify something first",
              "❌ No, let's start over"
            ]
          };
          setMessages(prev => prev.slice(0, -1).concat(confirmMessage));
        }
      } else {
        // If backend fails, show error message but don't generate fake AI responses
        const errorMessage: Message = {
          id: (Date.now() + 2).toString(),
          type: 'system',
          content: `❌ Connection error: ${response.error || 'Failed to get AI response'}. Please check your connection and try again.`,
          timestamp: new Date(),
          suggestions: [
            "🔄 Try again",
            "📞 Contact support",
            "🔧 Check connection"
          ]
        };
        setMessages(prev => prev.slice(0, -1).concat(errorMessage));
      }
    } catch (error) {
      // Network error - show system message, not fake AI response
      const errorMessage: Message = {
        id: (Date.now() + 2).toString(),
        type: 'system',
        content: `🌐 Network error: Unable to connect to AI assistant. Please check your internet connection and try again.`,
        timestamp: new Date(),
        suggestions: [
          "🔄 Retry message",
          "🔧 Check connection",
          "📞 Contact support"
        ]
      };
      setMessages(prev => prev.slice(0, -1).concat(errorMessage));
    } finally {
      setIsLoading(false);
    }
  };

  // Handle retry functionality for failed messages
  const retryLastMessage = () => {
    const lastUserMessage = messages.slice().reverse().find(m => m.type === 'user');
    if (lastUserMessage) {
      sendMessage(lastUserMessage.content);
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    // Handle special suggestion actions
    if (suggestion.includes('Upload') || suggestion.includes('upload')) {
      handleFileUpload();
      return;
    }
    if (suggestion.includes('template') || suggestion.includes('Template')) {
      if (suggestion.includes('student')) downloadTemplate('students');
      else if (suggestion.includes('teacher')) downloadTemplate('teachers');
      else if (suggestion.includes('parent')) downloadTemplate('parents');
      else downloadTemplate('students'); // default
      return;
    }
    
    sendMessage(suggestion);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage(inputValue);
    }
  };

  return (
    <div className="flex flex-col h-full bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden">
      {/* Enhanced Header */}
      <div className="flex items-center justify-between p-6 border-b border-gray-200 bg-gradient-to-r from-blue-50 via-indigo-50 to-purple-50">
        <div className="flex items-center space-x-4">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-600 via-indigo-600 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
            <Sparkles className="w-6 h-6 text-white" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">AI School Setup Assistant</h2>
            <p className="text-sm text-gray-600 flex items-center space-x-2">
              <span>🤖 Powered by Gemini AI</span>
              <span>•</span>
              <span>📊 Bulk Import</span>
              <span>•</span>
              <span>⚡ Smart Configuration</span>
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          {/* Enhanced Quick Action Buttons */}
          <button
            onClick={handleFileUpload}
            className="flex items-center space-x-2 px-4 py-2 bg-white border border-gray-300 rounded-xl hover:bg-blue-50 hover:border-blue-300 transition-all duration-200 shadow-sm"
            title="Upload Excel/CSV files for bulk import"
          >
            <Upload className="w-4 h-4 text-blue-600" />
            <span className="text-sm font-medium text-gray-700">Upload Files</span>
          </button>
          
          {setupComplete ? (
            <div className="flex items-center space-x-2 px-4 py-2 bg-green-50 border border-green-200 rounded-xl">
              <CheckCircle2 className="w-5 h-5 text-green-600" />
              <span className="text-sm font-semibold text-green-700">Setup Complete!</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2 px-4 py-2 bg-blue-50 border border-blue-200 rounded-xl">
              <MessageCircle className="w-5 h-5 text-blue-600 animate-pulse" />
              <span className="text-sm font-medium text-blue-700">Configuring...</span>
            </div>
          )}
        </div>
      </div>

      {/* Enhanced Messages Area */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6 bg-gradient-to-b from-gray-50 to-white">
        {!isInitialized && (
          <div className="flex items-center justify-center py-12">
            <div className="flex flex-col items-center space-y-4 text-gray-500">
              <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <Sparkles className="w-8 h-8 text-white animate-pulse" />
              </div>
              <div className="text-center">
                <div className="flex items-center space-x-2 text-lg font-medium">
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>Initializing AI assistant...</span>
                </div>
                <p className="text-sm text-gray-400 mt-1">Connecting to Gemini AI</p>
              </div>
            </div>
          </div>
        )}

        {messages.map((message, index) => (
          <div
            key={message.id}
            className={`flex items-start space-x-4 animate-fadeIn ${
              message.type === 'user' ? 'flex-row-reverse space-x-reverse' : ''
            }`}
            style={{ animationDelay: `${index * 0.1}s` }}
          >
            {/* Enhanced Avatar */}
            <div
              className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 shadow-lg ${
                message.type === 'user'
                  ? 'bg-gradient-to-br from-blue-500 to-blue-600'
                  : message.type === 'system'
                  ? 'bg-gradient-to-br from-yellow-500 to-orange-500'
                  : 'bg-gradient-to-br from-purple-500 to-indigo-600'
              }`}
            >
              {message.type === 'user' ? (
                <User className="w-5 h-5 text-white" />
              ) : message.type === 'system' ? (
                <Settings className="w-5 h-5 text-white" />
              ) : (
                <Bot className="w-5 h-5 text-white" />
              )}
            </div>

            {/* Enhanced Message Content */}
            <div
              className={`flex-1 max-w-[75%] ${
                message.type === 'user' ? 'text-right' : 'text-left'
              }`}
            >
              {/* Message Label */}
              <div className={`text-xs font-medium mb-2 ${
                message.type === 'user' ? 'text-blue-600' : message.type === 'system' ? 'text-yellow-600' : 'text-purple-600'
              }`}>
                {message.type === 'user' ? 'You' : message.type === 'system' ? 'System' : 'AI Assistant'}
                <span className="text-gray-400 ml-2">
                  {message.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>

              <div
                className={`inline-block px-5 py-4 rounded-2xl shadow-sm ${
                  message.type === 'user'
                    ? 'bg-gradient-to-br from-blue-500 to-blue-600 text-white'
                    : message.type === 'system'
                    ? 'bg-gradient-to-br from-yellow-50 to-orange-50 text-orange-800 border border-orange-200'
                    : 'bg-white text-gray-900 border border-gray-200 shadow-md'
                }`}
              >
                {message.isLoading ? (
                  <div className="flex items-center space-x-3">
                    <div className="flex space-x-1">
                      <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce"></div>
                      <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                      <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                    </div>
                    <span className="text-sm font-medium">AI is analyzing your request...</span>
                  </div>
                ) : (
                  <div className="whitespace-pre-wrap text-sm leading-relaxed">{message.content}</div>
                )}
              </div>

              {/* Attachments */}
              {message.attachments && message.attachments.length > 0 && (
                <div className="mt-2 space-y-2">
                  {message.attachments.map((attachment, index) => (
                    <div
                      key={index}
                      className="flex items-center space-x-2 p-2 bg-gray-50 rounded-lg border"
                    >
                      <FileSpreadsheet className="w-4 h-4 text-green-600" />
                      <div className="flex-1">
                        <div className="text-sm font-medium text-gray-900">{attachment.name}</div>
                        {attachment.description && (
                          <div className="text-xs text-gray-500">{attachment.description}</div>
                        )}
                      </div>
                      {attachment.downloadUrl && (
                        <button
                          onClick={() => window.open(attachment.downloadUrl, '_blank')}
                          className="p-1 text-blue-600 hover:bg-blue-50 rounded"
                        >
                          <Download className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {/* Enhanced Suggestions */}
              {message.suggestions && message.suggestions.length > 0 && !message.isLoading && (
                <div className="mt-4">
                  <div className="text-xs font-medium text-gray-500 mb-2">💡 Quick suggestions:</div>
                  <div className="flex flex-wrap gap-2">
                    {message.suggestions.map((suggestion, index) => (
                      <button
                        key={index}
                        onClick={() => handleSuggestionClick(suggestion)}
                        className="px-4 py-2 text-sm bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-xl hover:from-blue-100 hover:to-indigo-100 hover:border-blue-300 transition-all duration-200 text-left text-blue-800 font-medium shadow-sm hover:shadow-md"
                      >
                        {suggestion}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Enhanced Input Area */}
      <div className="border-t border-gray-200 bg-gradient-to-r from-gray-50 to-white p-6">
        <div className="flex items-end space-x-4">
          <div className="flex-1 relative">
            <div className="absolute left-4 top-4 text-gray-400">
              <MessageCircle className="w-5 h-5" />
            </div>
            <textarea
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Tell me about your school, upload data files, or ask for help with setup..."
              className="w-full pl-12 pr-16 py-4 border-2 border-gray-200 rounded-2xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none transition-all duration-200 shadow-sm"
              disabled={isLoading}
              rows={inputValue.length > 50 ? 3 : 1}
            />
            <div className="absolute right-2 bottom-2 flex items-center space-x-2">
              <button
                onClick={handleFileUpload}
                className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-all duration-200"
                title="Upload Excel/CSV files"
              >
                <Upload className="w-5 h-5" />
              </button>
              <button
                onClick={() => sendMessage(inputValue)}
                disabled={!inputValue.trim() || isLoading}
                className="p-2 bg-blue-600 text-white hover:bg-blue-700 rounded-lg disabled:bg-gray-300 disabled:cursor-not-allowed transition-all duration-200 shadow-sm hover:shadow-md"
              >
                {isLoading ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : (
                  <Send className="w-5 h-5" />
                )}
              </button>
            </div>
          </div>
        </div>
        
        {/* Enhanced Upload Progress */}
        {uploadedFiles.length > 0 && (
          <div className="mt-4 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-xl">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
                <FileSpreadsheet className="w-4 h-4 text-white" />
              </div>
              <div>
                <div className="text-sm font-semibold text-blue-900">
                  📁 {uploadedFiles.length} file(s) ready for processing
                </div>
                <div className="text-xs text-blue-600">
                  Files will be processed when you send your next message
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Helpful Tips */}
        {messages.length === 1 && (
          <div className="mt-4 p-4 bg-gradient-to-r from-purple-50 to-pink-50 border border-purple-200 rounded-xl">
            <div className="text-sm text-purple-800">
              <div className="font-semibold mb-2">💡 Pro Tips:</div>
              <ul className="space-y-1 text-xs">
                <li>• Be specific about your school's needs and requirements</li>
                <li>• Upload Excel/CSV files for bulk data import</li>
                <li>• Ask about features like attendance, grading, or parent communication</li>
                <li>• The AI will guide you through the entire setup process</li>
              </ul>
            </div>
          </div>
        )}
      </div>

      {/* File Upload Modal */}
      {showFileUpload && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Upload School Data</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Select files to upload
                </label>
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  accept=".xlsx,.xls,.csv"
                  onChange={handleFileSelect}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div className="text-sm text-gray-600">
                Supported formats: Excel (.xlsx, .xls), CSV (.csv)
              </div>
              <div className="flex space-x-3">
                <button
                  onClick={() => setShowFileUpload(false)}
                  className="flex-1 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={() => downloadTemplate('students')}
                  className="flex-1 px-4 py-2 text-blue-700 bg-blue-100 rounded-lg hover:bg-blue-200 transition-colors"
                >
                  Get Template
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AISchoolSetupChat; 