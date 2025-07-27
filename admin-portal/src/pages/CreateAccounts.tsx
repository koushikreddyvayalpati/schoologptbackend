import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { 
  UserPlus, 
  Upload, 
  Download, 
  Users, 
  GraduationCap, 
  BookOpen,
  Mail,
  User,
  Key,
  Building,
  FileText,
  CheckCircle,
  AlertCircle,
  X,
  Plus,
  Check
} from 'lucide-react';

interface CreateUserFormData {
  name: string;
  email: string;
  role: 'teacher' | 'student';
  grade?: string;
  subject?: string;
  department?: string;
  studentId?: string;
  employeeId?: string;
}

const CreateAccounts: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'single' | 'bulk'>('single');
  const [createdUsers, setCreatedUsers] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors, isValid },
  } = useForm<CreateUserFormData>({
    mode: 'onChange',
  });

  const selectedRole = watch('role');

  const onSubmit = async (data: CreateUserFormData) => {
    setIsLoading(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newUser = {
        id: Date.now(),
        ...data,
        password: generateRandomPassword(),
        createdAt: new Date().toISOString(),
      };
      
      setCreatedUsers([...createdUsers, newUser]);
      reset();
    } catch (error) {
      console.error('Failed to create user:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const generateRandomPassword = () => {
    return Math.random().toString(36).slice(-8);
  };

  const downloadTemplate = () => {
    const csvContent = `Name,Email,Role,Grade,Subject,Department,Student ID,Employee ID
John Doe,john.doe@school.edu,student,9,,,ST001,
Jane Smith,jane.smith@school.edu,teacher,,Mathematics,Science Dept,,EMP001`;
    
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'user_import_template.csv';
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const handleBulkImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // Handle CSV file parsing here
              // File import processing
    }
  };

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900">
            Create User Accounts
          </h1>
          <p className="text-gray-600 text-lg">
            Add new teachers and students to your school system
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <button
            onClick={downloadTemplate}
            className="btn-secondary"
          >
            <Download className="w-4 h-4 mr-2" />
            Download Template
          </button>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="card">
        <div className="border-b border-gray-200">
          <nav className="flex space-x-8 px-6">
            <button
              onClick={() => setActiveTab('single')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'single'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center space-x-2">
                <UserPlus className="w-4 h-4" />
                <span>Single User</span>
              </div>
            </button>
            <button
              onClick={() => setActiveTab('bulk')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'bulk'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center space-x-2">
                <Upload className="w-4 h-4" />
                <span>Bulk Import</span>
              </div>
            </button>
          </nav>
        </div>

        {/* Single User Form */}
        {activeTab === 'single' && (
          <div className="p-6">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* Role Selection */}
              <div className="space-y-3">
                <label className="text-sm font-semibold text-gray-700">
                  User Role *
                </label>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <label className="relative">
                    <input
                      {...register('role', { required: 'Role is required' })}
                      type="radio"
                      value="teacher"
                      className="sr-only"
                    />
                    <div className={`p-4 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedRole === 'teacher'
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}>
                      <div className="flex items-center space-x-3">
                        <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                          selectedRole === 'teacher'
                            ? 'bg-blue-100 text-blue-600'
                            : 'bg-gray-100 text-gray-400'
                        }`}>
                          <BookOpen className="w-5 h-5" />
                        </div>
                        <div>
                          <h3 className="font-medium text-gray-900">Teacher</h3>
                          <p className="text-sm text-gray-500">Teaching staff member</p>
                        </div>
                      </div>
                    </div>
                  </label>

                  <label className="relative">
                    <input
                      {...register('role', { required: 'Role is required' })}
                      type="radio"
                      value="student"
                      className="sr-only"
                    />
                    <div className={`p-4 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedRole === 'student'
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}>
                      <div className="flex items-center space-x-3">
                        <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                          selectedRole === 'student'
                            ? 'bg-blue-100 text-blue-600'
                            : 'bg-gray-100 text-gray-400'
                        }`}>
                          <GraduationCap className="w-5 h-5" />
                        </div>
                        <div>
                          <h3 className="font-medium text-gray-900">Student</h3>
                          <p className="text-sm text-gray-500">Enrolled learner</p>
                        </div>
                      </div>
                    </div>
                  </label>
                </div>
                {errors.role && (
                  <p className="text-red-600 text-sm flex items-center space-x-1">
                    <AlertCircle className="h-4 w-4" />
                    <span>{errors.role.message}</span>
                  </p>
                )}
              </div>

              {/* Basic Information */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label htmlFor="name" className="text-sm font-semibold text-gray-700">
                    Full Name *
                  </label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <User className="h-5 w-5 text-gray-400" />
                    </div>
                    <input
                      {...register('name', { required: 'Name is required' })}
                      type="text"
                      id="name"
                      className={`input pl-10 ${errors.name ? 'input-error' : ''}`}
                      placeholder="Enter full name"
                    />
                  </div>
                  {errors.name && (
                    <p className="text-red-600 text-sm flex items-center space-x-1">
                      <AlertCircle className="h-4 w-4" />
                      <span>{errors.name.message}</span>
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="email" className="text-sm font-semibold text-gray-700">
                    Email Address *
                  </label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <Mail className="h-5 w-5 text-gray-400" />
                    </div>
                    <input
                      {...register('email', {
                        required: 'Email is required',
                        pattern: {
                          value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                          message: 'Invalid email address',
                        },
                      })}
                      type="email"
                      id="email"
                      className={`input pl-10 ${errors.email ? 'input-error' : ''}`}
                      placeholder="Enter email address"
                    />
                  </div>
                  {errors.email && (
                    <p className="text-red-600 text-sm flex items-center space-x-1">
                      <AlertCircle className="h-4 w-4" />
                      <span>{errors.email.message}</span>
                    </p>
                  )}
                </div>
              </div>

              {/* Role-specific fields */}
              {selectedRole === 'teacher' && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <label htmlFor="subject" className="text-sm font-semibold text-gray-700">
                      Subject/Course
                    </label>
                    <input
                      {...register('subject')}
                      type="text"
                      id="subject"
                      className="input"
                      placeholder="e.g., Mathematics, English"
                    />
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="department" className="text-sm font-semibold text-gray-700">
                      Department
                    </label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Building className="h-5 w-5 text-gray-400" />
                      </div>
                      <input
                        {...register('department')}
                        type="text"
                        id="department"
                        className="input pl-10"
                        placeholder="e.g., Science Department"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="employeeId" className="text-sm font-semibold text-gray-700">
                      Employee ID
                    </label>
                    <input
                      {...register('employeeId')}
                      type="text"
                      id="employeeId"
                      className="input"
                      placeholder="e.g., EMP001"
                    />
                  </div>
                </div>
              )}

              {selectedRole === 'student' && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <label htmlFor="grade" className="text-sm font-semibold text-gray-700">
                      Grade/Class *
                    </label>
                    <select
                      {...register('grade', { required: selectedRole === 'student' ? 'Grade is required' : false })}
                      id="grade"
                      className={`input ${errors.grade ? 'input-error' : ''}`}
                    >
                      <option value="">Select Grade</option>
                      <option value="9">Grade 9</option>
                      <option value="10">Grade 10</option>
                      <option value="11">Grade 11</option>
                      <option value="12">Grade 12</option>
                    </select>
                    {errors.grade && (
                      <p className="text-red-600 text-sm flex items-center space-x-1">
                        <AlertCircle className="h-4 w-4" />
                        <span>{errors.grade.message}</span>
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="studentId" className="text-sm font-semibold text-gray-700">
                      Student ID
                    </label>
                    <input
                      {...register('studentId')}
                      type="text"
                      id="studentId"
                      className="input"
                      placeholder="e.g., ST001"
                    />
                  </div>
                </div>
              )}

              {/* Submit Button */}
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={!isValid || isLoading}
                  className="btn-primary px-8"
                >
                  {isLoading ? (
                    <div className="flex items-center">
                      <div className="spinner w-4 h-4 mr-2"></div>
                      Creating...
                    </div>
                  ) : (
                    <div className="flex items-center">
                      <UserPlus className="w-4 h-4 mr-2" />
                      Create Account
                    </div>
                  )}
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Bulk Import */}
        {activeTab === 'bulk' && (
          <div className="p-6">
            <div className="space-y-6">
              {/* Upload Area */}
              <div className="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center hover:border-blue-400 transition-colors">
                <div className="space-y-4">
                  <div className="w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto">
                    <Upload className="w-8 h-8 text-blue-600" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">Upload CSV File</h3>
                    <p className="text-gray-600">
                      Import multiple users at once using a CSV file
                    </p>
                  </div>
                  <div className="space-y-2">
                    <input
                      type="file"
                      accept=".csv"
                      onChange={handleBulkImport}
                      className="hidden"
                      id="bulk-upload"
                    />
                    <label
                      htmlFor="bulk-upload"
                      className="btn-primary cursor-pointer inline-flex"
                    >
                      <Upload className="w-4 h-4 mr-2" />
                      Choose CSV File
                    </label>
                    <p className="text-sm text-gray-500">
                      Maximum file size: 5MB
                    </p>
                  </div>
                </div>
              </div>

              {/* Instructions */}
              <div className="bg-blue-50 border border-blue-200 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-blue-900 mb-4">
                  CSV Format Instructions
                </h3>
                <div className="space-y-3 text-sm text-blue-800">
                  <p>Your CSV file should include the following columns:</p>
                  <ul className="list-disc list-inside space-y-1 ml-4">
                    <li><strong>Name:</strong> Full name of the user</li>
                    <li><strong>Email:</strong> Valid email address</li>
                    <li><strong>Role:</strong> Either "teacher" or "student"</li>
                    <li><strong>Grade:</strong> Required for students (9, 10, 11, 12)</li>
                    <li><strong>Subject:</strong> Optional for teachers</li>
                    <li><strong>Department:</strong> Optional for teachers</li>
                    <li><strong>Student ID:</strong> Optional for students</li>
                    <li><strong>Employee ID:</strong> Optional for teachers</li>
                  </ul>
                  <div className="mt-4">
                    <button
                      onClick={downloadTemplate}
                      className="btn-secondary text-sm"
                    >
                      <Download className="w-4 h-4 mr-2" />
                      Download Template
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Created Users List */}
      {createdUsers.length > 0 && (
        <div className="card animate-slide-up">
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">
                Recently Created Accounts
              </h2>
              <span className="badge-success">
                {createdUsers.length} users created
              </span>
            </div>
            <div className="space-y-4">
              {createdUsers.map((user) => (
                <div
                  key={user.id}
                  className="flex items-center justify-between p-4 bg-green-50 border border-green-200 rounded-xl"
                >
                  <div className="flex items-center space-x-4">
                    <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                      user.role === 'teacher' ? 'bg-blue-100 text-blue-600' : 'bg-purple-100 text-purple-600'
                    }`}>
                      {user.role === 'teacher' ? (
                        <BookOpen className="w-5 h-5" />
                      ) : (
                        <GraduationCap className="w-5 h-5" />
                      )}
                    </div>
                    <div>
                      <h3 className="font-medium text-gray-900">{user.name}</h3>
                      <p className="text-sm text-gray-600">{user.email}</p>
                    </div>
                    <span className={`badge ${
                      user.role === 'teacher' ? 'badge-info' : 'badge-success'
                    }`}>
                      {user.role}
                    </span>
                  </div>
                  <div className="flex items-center space-x-4">
                    <div className="text-right">
                      <p className="text-sm font-medium text-gray-900">Password: {user.password}</p>
                      <p className="text-xs text-gray-500">Temporary password</p>
                    </div>
                    <CheckCircle className="w-5 h-5 text-green-500" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CreateAccounts; 