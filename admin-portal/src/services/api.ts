// API base URL - update this to match your backend
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// API response types
interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

interface SchoolSetupRequest {
  schoolName: string;
  adminName: string;
  adminEmail: string;
  description?: string;
  address?: string;
  phone?: string;
}

interface ChatMessage {
  message: string;
  schoolName?: string;
  adminName?: string;
  adminEmail?: string;
}

interface ChatResponse {
  response?: string;
  suggestions?: string[];
  next_step?: string;
  agent_response?: string;
  session_status?: string;
  confirmation_pending?: boolean;
  generated_config?: {
    school_name: string;
    region: string;
    education_system: string;
    features: Record<string, boolean>;
  };
  message_count?: number;
  session_id?: string;
  welcome_message?: string;
}

interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'teacher' | 'student' | 'parent';
  schoolId?: string;
  avatar?: string;
  emailVerified?: boolean;
}

interface School {
  id: string;
  school_name: string;
  school_code: string;
  admin_email: string;
  region: string;
  education_system: string;
  timezone: string;
  language: string;
  currency: string;
  status: string;
  created_at: string;
  updated_at: string;
  features: {
    attendance_tracking: boolean;
    grade_management: boolean;
    parent_communication: boolean;
    voice_processing: boolean;
    advanced_analytics: boolean;
    ai_insights: boolean;
    online_exams: boolean;
    multi_language_support: boolean;
  };
}

class ApiService {
  private baseURL: string;
  private authToken: string | null = null;

  constructor() {
    this.baseURL = API_BASE_URL;
    // Try to get token from localStorage
    this.authToken = localStorage.getItem('authToken');
  }

  // Set authentication token
  setAuthToken(token: string) {
    this.authToken = token;
    localStorage.setItem('authToken', token);
  }

  // Clear authentication token
  clearAuthToken() {
    this.authToken = null;
    localStorage.removeItem('authToken');
  }

  // Generic API request method
  private async request<T = any>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    // Add auth token if available
    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        return {
          success: false,
          error: data.error || `HTTP ${response.status}: ${response.statusText}`,
        };
      }

      return {
        success: true,
        data,
      };
    } catch (error) {
      console.error('API request failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }

  // Health check
  async healthCheck(): Promise<ApiResponse> {
    return this.request('/health');
  }

  // School Setup APIs
  async startSchoolSetup(data: SchoolSetupRequest): Promise<ApiResponse> {
    return this.request('/api/v1/setup/school', {
      method: 'POST',
      body: JSON.stringify({
        school_name: data.schoolName,
        admin_email: data.adminEmail,
        region: data.address || "Unknown", // Use address as region for now
        education_system: "cbse", // Default education system
        timezone: "UTC", // Default timezone
        language: "en", // Default language
        currency: "USD", // Default currency
        features: {
          attendance_tracking: true,
          grade_management: true,
          parent_communication: true,
          voice_processing: false,
          advanced_analytics: false,
          ai_insights: false,
          online_exams: false,
          multi_language_support: false,
        }
      }),
    });
  }

  async getSchoolsByAdmin(adminEmail: string): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/schools?admin_email=${encodeURIComponent(adminEmail)}`);
  }

  async getSchoolSetupTemplates(): Promise<ApiResponse> {
    return this.request('/api/v1/setup/templates');
  }

  async getSchoolConfiguration(schoolId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/school/${schoolId}`);
  }

  async getSetupProgress(schoolId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/school/${schoolId}/progress`);
  }

  async updateSetupProgress(schoolId: string, progress: any): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/school/${schoolId}/progress`, {
      method: 'PUT',
      body: JSON.stringify(progress),
    });
  }

  // AI Chat APIs
  async startSetupChat(data: { admin_name: string; admin_email: string }): Promise<ApiResponse<ChatResponse>> {
    return this.request('/api/v1/setup/chat/start', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async sendChatMessage(sessionId: string, message: string): Promise<ApiResponse<ChatResponse>> {
    return this.request(`/api/v1/setup/chat/${sessionId}/message`, {
      method: 'POST',
      body: JSON.stringify({ message }),
    });
  }

  async getChatHistory(sessionId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/chat/${sessionId}/history`);
  }

  async confirmConfiguration(sessionId: string, message?: string): Promise<ApiResponse> {
    return this.request(`/api/v1/setup/chat/${sessionId}/confirm`, {
      method: 'POST',
      body: JSON.stringify({ 
        confirmed: true,
        message: message || "Yes, create the school with the selected features"
      }),
    });
  }

  async getChatExamples(): Promise<ApiResponse> {
    return this.request('/api/v1/setup/chat/examples');
  }

  async getActiveSessions(): Promise<ApiResponse> {
    return this.request('/api/v1/setup/chat/sessions');
  }

  // GPT APIs
  async getGPTResponse(query: string): Promise<ApiResponse<{ answer: string; model: string }>> {
    return this.request('/test-gpt', {
      method: 'POST',
      body: JSON.stringify({ query }),
    });
  }

  async askGPT(query: string): Promise<ApiResponse> {
    return this.request('/api/v1/gpt/ask', {
      method: 'POST',
      body: JSON.stringify({ query }),
    });
  }

  async askGPTAttendance(query: string): Promise<ApiResponse> {
    return this.request('/api/v1/gpt/attendance', {
      method: 'POST',
      body: JSON.stringify({ query }),
    });
  }

  // User Management APIs
  async createUser(userData: {
    name: string;
    email: string;
    role: 'teacher' | 'student' | 'parent';
    grade?: string;
    subject?: string;
    department?: string;
    studentId?: string;
    employeeId?: string;
    phone?: string;
    address?: string;
  }): Promise<ApiResponse<User>> {
    return this.request('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify({
        name: userData.name,
        email: userData.email,
        role: userData.role,
        grade: userData.grade,
        subject: userData.subject,
        department: userData.department,
        student_id: userData.studentId,
        employee_id: userData.employeeId,
        phone: userData.phone,
        address: userData.address,
      }),
    });
  }

  async bulkCreateUsers(users: Array<{
    name: string;
    email: string;
    role: 'teacher' | 'student' | 'parent';
    grade?: string;
    subject?: string;
    department?: string;
    studentId?: string;
    employeeId?: string;
    phone?: string;
    address?: string;
  }>): Promise<ApiResponse<{
    created_count: number;
    failed_count: number;
    created_users: User[];
    failed_users?: Array<{
      index: number;
      email: string;
      error: string;
    }>;
  }>> {
    const formattedUsers = users.map(user => ({
      name: user.name,
      email: user.email,
      role: user.role,
      grade: user.grade,
      subject: user.subject,
      department: user.department,
      student_id: user.studentId,
      employee_id: user.employeeId,
      phone: user.phone,
      address: user.address,
    }));

    return this.request('/api/v1/users/bulk', {
      method: 'POST',
      body: JSON.stringify({
        users: formattedUsers,
      }),
    });
  }

  async uploadCSV(file: File): Promise<ApiResponse<{
    created_count: number;
    failed_count: number;
    created_users: User[];
    failed_users?: Array<{
      row: number;
      email: string;
      error: string;
    }>;
  }>> {
    const formData = new FormData();
    formData.append('csv_file', file);

    return this.request('/api/v1/users/upload-csv', {
      method: 'POST',
      body: formData,
      headers: {
        // Don't set Content-Type, let browser set it for FormData
      },
    });
  }

  async getUsers(filters?: {
    role?: string;
    grade?: string;
    department?: string;
  }): Promise<ApiResponse<{
    count: number;
    users: User[];
  }>> {
    const params = new URLSearchParams();
    if (filters?.role) params.append('role', filters.role);
    if (filters?.grade) params.append('grade', filters.grade);
    if (filters?.department) params.append('department', filters.department);

    const queryString = params.toString();
    const endpoint = queryString ? `/api/v1/users?${queryString}` : '/api/v1/users';

    return this.request(endpoint);
  }

  async getUserById(userId: string): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${userId}`);
  }

  async updateUser(userId: string, userData: Partial<User>): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }

  async deleteUser(userId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/users/${userId}`, {
      method: 'DELETE',
    });
  }

  // Educational APIs
  async createSubject(subjectData: { name: string; description?: string; grade: string }): Promise<ApiResponse> {
    return this.request('/api/v1/education/subjects', {
      method: 'POST',
      body: JSON.stringify(subjectData),
    });
  }

  async createAssignment(assignmentData: {
    title: string;
    description: string;
    subject_id: string;
    due_date: string;
    max_points: number;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/education/assignments', {
      method: 'POST',
      body: JSON.stringify(assignmentData),
    });
  }

  async getTeacherAssignments(): Promise<ApiResponse> {
    return this.request('/api/v1/education/assignments/teacher');
  }

  async getAssignmentSubmissions(assignmentId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/education/assignments/${assignmentId}/submissions`);
  }

  async submitAssignment(submissionData: {
    assignment_id: string;
    content: string;
    file_url?: string;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/education/assignments/submit', {
      method: 'POST',
      body: JSON.stringify(submissionData),
    });
  }

  async gradeSubmission(submissionId: string, gradeData: {
    score: number;
    feedback?: string;
  }): Promise<ApiResponse> {
    return this.request(`/api/v1/education/submissions/${submissionId}/grade`, {
      method: 'POST',
      body: JSON.stringify(gradeData),
    });
  }

  async getStudentGrades(studentId: string): Promise<ApiResponse> {
    return this.request(`/api/v1/education/students/${studentId}/grades`);
  }

  async getStudentDashboard(): Promise<ApiResponse> {
    return this.request('/api/v1/education/dashboard/student');
  }

  async getTeacherDashboard(): Promise<ApiResponse> {
    return this.request('/api/v1/education/dashboard/teacher');
  }

  // Attendance APIs
  async getAttendance(studentId: string, date: string): Promise<ApiResponse> {
    return this.request(`/api/v1/attendance/${studentId}/${date}`);
  }

  async createAttendance(attendanceData: {
    student_id: string;
    date: string;
    status: 'present' | 'absent' | 'late';
    notes?: string;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/attendance', {
      method: 'POST',
      body: JSON.stringify(attendanceData),
    });
  }

  // Teacher Tasks APIs
  async markAttendance(attendanceData: {
    class_id: string;
    date: string;
    attendances: Array<{
      student_id: string;
      status: 'present' | 'absent' | 'late';
    }>;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/teacher/mark-attendance', {
      method: 'POST',
      body: JSON.stringify(attendanceData),
    });
  }

  async analyzeStudent(studentData: {
    student_id: string;
    analysis_type: string;
    context?: string;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/teacher/analyze-student', {
      method: 'POST',
      body: JSON.stringify(studentData),
    });
  }

  async notifyParent(notificationData: {
    student_id: string;
    message: string;
    type: string;
  }): Promise<ApiResponse> {
    return this.request('/api/v1/teacher/notify-parent', {
      method: 'POST',
      body: JSON.stringify(notificationData),
    });
  }

  // Voice APIs
  async transcribeAudio(audioData: Blob): Promise<ApiResponse> {
    const formData = new FormData();
    formData.append('audio', audioData);

    return this.request('/api/v1/voice/transcribe', {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set content-type for FormData
    });
  }

  async synthesizeSpeech(textData: { text: string; voice?: string }): Promise<ApiResponse> {
    return this.request('/api/v1/voice/synthesize', {
      method: 'POST',
      body: JSON.stringify(textData),
    });
  }
}

// Create and export a singleton instance
const apiService = new ApiService();
export default apiService;

// Export types for use in components
export type {
  ApiResponse,
  SchoolSetupRequest,
  ChatMessage,
  ChatResponse,
  User,
  School,
}; 