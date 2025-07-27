import React, { useState } from 'react';
import { 
  Users, 
  UserPlus, 
  Search, 
  Filter, 
  MoreVertical,
  Calendar,
  Clock,
  GraduationCap,
  BookOpen,
  CheckCircle,
  X,
  Plus,
  Edit,
  Trash2,
  Mail,
  Phone,
  MapPin,
  Award,
  TrendingUp,
  User
} from 'lucide-react';

interface Student {
  id: string;
  name: string;
  email: string;
  grade: string;
  studentId: string;
  avatar?: string;
  attendance: number;
  gpa: number;
  lastActivity: string;
  status: 'active' | 'inactive';
}

interface Class {
  id: string;
  name: string;
  subject: string;
  grade: string;
  schedule: string;
  room: string;
  students: Student[];
  capacity: number;
}

const TeacherClasses: React.FC = () => {
  const [selectedClass, setSelectedClass] = useState<string>('class1');
  const [searchTerm, setSearchTerm] = useState('');
  const [showAddStudent, setShowAddStudent] = useState(false);
  const [selectedStudents, setSelectedStudents] = useState<string[]>([]);

  // Mock data
  const classes: Class[] = [
    {
      id: 'class1',
      name: 'Mathematics 10A',
      subject: 'Mathematics',
      grade: '10',
      schedule: 'Mon, Wed, Fri - 9:00 AM',
      room: 'Room 201',
      capacity: 30,
      students: [
        {
          id: 'st1',
          name: 'Alice Johnson',
          email: 'alice.johnson@school.edu',
          grade: '10',
          studentId: 'ST001',
          attendance: 95,
          gpa: 3.8,
          lastActivity: '2 hours ago',
          status: 'active'
        },
        {
          id: 'st2',
          name: 'Bob Smith',
          email: 'bob.smith@school.edu',
          grade: '10',
          studentId: 'ST002',
          attendance: 87,
          gpa: 3.2,
          lastActivity: '1 day ago',
          status: 'active'
        },
        {
          id: 'st3',
          name: 'Carol Davis',
          email: 'carol.davis@school.edu',
          grade: '10',
          studentId: 'ST003',
          attendance: 92,
          gpa: 3.9,
          lastActivity: '3 hours ago',
          status: 'active'
        }
      ]
    },
    {
      id: 'class2',
      name: 'Mathematics 10B',
      subject: 'Mathematics',
      grade: '10',
      schedule: 'Tue, Thu - 10:30 AM',
      room: 'Room 202',
      capacity: 25,
      students: [
        {
          id: 'st4',
          name: 'David Wilson',
          email: 'david.wilson@school.edu',
          grade: '10',
          studentId: 'ST004',
          attendance: 89,
          gpa: 3.5,
          lastActivity: '5 hours ago',
          status: 'active'
        }
      ]
    },
    {
      id: 'class3',
      name: 'Algebra 11C',
      subject: 'Mathematics',
      grade: '11',
      schedule: 'Mon, Wed, Fri - 2:00 PM',
      room: 'Room 203',
      capacity: 28,
      students: []
    }
  ];

  const availableStudents = [
    { id: 'av1', name: 'Emma Brown', grade: '10', studentId: 'ST005' },
    { id: 'av2', name: 'James Miller', grade: '10', studentId: 'ST006' },
    { id: 'av3', name: 'Sarah Taylor', grade: '10', studentId: 'ST007' },
    { id: 'av4', name: 'Michael Johnson', grade: '10', studentId: 'ST008' },
  ];

  const currentClass = classes.find(c => c.id === selectedClass);

  const filteredStudents = currentClass?.students.filter(student =>
    student.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    student.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
    student.studentId.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

      const handleAddStudents = () => {
      // Add selected students to class
      setShowAddStudent(false);
      setSelectedStudents([]);
  };

  const toggleStudentSelection = (studentId: string) => {
    setSelectedStudents(prev =>
      prev.includes(studentId)
        ? prev.filter(id => id !== studentId)
        : [...prev, studentId]
    );
  };

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900">
            My Classes
          </h1>
          <p className="text-gray-600 text-lg">
            Manage your classes and students
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <button
            onClick={() => setShowAddStudent(true)}
            className="btn-primary"
          >
            <UserPlus className="w-4 h-4 mr-2" />
            Add Students
          </button>
        </div>
      </div>

      {/* Class Tabs */}
      <div className="card">
        <div className="border-b border-gray-200">
          <nav className="flex space-x-8 px-6 overflow-x-auto">
            {classes.map((classItem) => (
              <button
                key={classItem.id}
                onClick={() => setSelectedClass(classItem.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap transition-colors ${
                  selectedClass === classItem.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <div className="flex items-center space-x-2">
                  <BookOpen className="w-4 h-4" />
                  <span>{classItem.name}</span>
                  <span className="bg-gray-100 text-gray-600 px-2 py-1 rounded-full text-xs">
                    {classItem.students.length}/{classItem.capacity}
                  </span>
                </div>
              </button>
            ))}
          </nav>
        </div>

        {/* Class Overview */}
        {currentClass && (
          <div className="p-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
              {/* Class Info */}
              <div className="lg:col-span-2">
                <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-2xl p-6 border border-blue-200">
                  <div className="flex items-start justify-between">
                    <div className="space-y-4">
                      <div>
                        <h2 className="text-2xl font-bold text-gray-900">{currentClass.name}</h2>
                        <p className="text-gray-600">{currentClass.subject} • Grade {currentClass.grade}</p>
                      </div>
                      <div className="flex items-center space-x-6 text-sm text-gray-600">
                        <div className="flex items-center space-x-2">
                          <Clock className="w-4 h-4" />
                          <span>{currentClass.schedule}</span>
                        </div>
                        <div className="flex items-center space-x-2">
                          <MapPin className="w-4 h-4" />
                          <span>{currentClass.room}</span>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-blue-600">
                        {currentClass.students.length}
                      </div>
                      <div className="text-sm text-gray-500">
                        of {currentClass.capacity} students
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Quick Stats */}
              <div className="space-y-4">
                <div className="card p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500">Avg Attendance</p>
                      <p className="text-2xl font-bold text-green-600">
                        {currentClass.students.length > 0 
                          ? Math.round(currentClass.students.reduce((acc, s) => acc + s.attendance, 0) / currentClass.students.length)
                          : 0}%
                      </p>
                    </div>
                    <CheckCircle className="w-8 h-8 text-green-500" />
                  </div>
                </div>
                <div className="card p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500">Avg GPA</p>
                      <p className="text-2xl font-bold text-purple-600">
                        {currentClass.students.length > 0 
                          ? (currentClass.students.reduce((acc, s) => acc + s.gpa, 0) / currentClass.students.length).toFixed(1)
                          : 0}
                      </p>
                    </div>
                    <Award className="w-8 h-8 text-purple-500" />
                  </div>
                </div>
              </div>
            </div>

            {/* Search and Filter */}
            <div className="flex flex-col sm:flex-row gap-4 mb-6">
              <div className="relative flex-1">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Search className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  placeholder="Search students..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="input pl-10"
                />
              </div>
              <button className="btn-secondary">
                <Filter className="w-4 h-4 mr-2" />
                Filter
              </button>
            </div>

            {/* Students List */}
            <div className="space-y-4">
              {filteredStudents.length === 0 ? (
                <div className="text-center py-12">
                  <div className="w-16 h-16 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                    <Users className="w-8 h-8 text-gray-400" />
                  </div>
                  <h3 className="text-lg font-medium text-gray-900">No students found</h3>
                  <p className="text-gray-500 mt-2">
                    {searchTerm ? 'Try adjusting your search terms' : 'Add students to get started'}
                  </p>
                  <button
                    onClick={() => setShowAddStudent(true)}
                    className="btn-primary mt-4"
                  >
                    <UserPlus className="w-4 h-4 mr-2" />
                    Add Students
                  </button>
                </div>
              ) : (
                filteredStudents.map((student) => (
                  <div
                    key={student.id}
                    className="card p-6 hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-4">
                        <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center text-white font-semibold">
                          {student.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div className="space-y-1">
                          <div className="flex items-center space-x-3">
                            <h3 className="font-semibold text-gray-900">{student.name}</h3>
                            <span className="badge badge-info">{student.studentId}</span>
                            <span className={`badge ${
                              student.status === 'active' ? 'badge-success' : 'badge-secondary'
                            }`}>
                              {student.status}
                            </span>
                          </div>
                          <div className="flex items-center space-x-4 text-sm text-gray-500">
                            <div className="flex items-center space-x-1">
                              <Mail className="w-4 h-4" />
                              <span>{student.email}</span>
                            </div>
                            <div className="flex items-center space-x-1">
                              <Clock className="w-4 h-4" />
                              <span>Active {student.lastActivity}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center space-x-6">
                        <div className="text-center">
                          <div className="text-lg font-semibold text-green-600">
                            {student.attendance}%
                          </div>
                          <div className="text-xs text-gray-500">Attendance</div>
                        </div>
                        <div className="text-center">
                          <div className="text-lg font-semibold text-purple-600">
                            {student.gpa}
                          </div>
                          <div className="text-xs text-gray-500">GPA</div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <button className="btn-ghost p-2">
                            <Edit className="w-4 h-4" />
                          </button>
                          <button className="btn-ghost p-2">
                            <MoreVertical className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>

      {/* Add Students Modal */}
      {showAddStudent && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">
                Add Students to {currentClass?.name}
              </h2>
              <button
                onClick={() => setShowAddStudent(false)}
                className="btn-ghost p-2"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            
            <div className="p-6 overflow-y-auto max-h-[60vh]">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <p className="text-gray-600">
                    Select students to add to your class
                  </p>
                  <span className="text-sm text-gray-500">
                    {selectedStudents.length} selected
                  </span>
                </div>
                
                {availableStudents.map((student) => (
                  <div
                    key={student.id}
                    className={`p-4 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedStudents.includes(student.id)
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                    onClick={() => toggleStudentSelection(student.id)}
                  >
                    <div className="flex items-center space-x-4">
                      <div className={`w-6 h-6 rounded-full border-2 flex items-center justify-center ${
                        selectedStudents.includes(student.id)
                          ? 'border-blue-500 bg-blue-500'
                          : 'border-gray-300'
                      }`}>
                        {selectedStudents.includes(student.id) && (
                          <CheckCircle className="w-4 h-4 text-white" />
                        )}
                      </div>
                      <div className="w-10 h-10 bg-gradient-to-br from-green-500 to-blue-600 rounded-xl flex items-center justify-center text-white font-semibold">
                        {student.name.split(' ').map(n => n[0]).join('')}
                      </div>
                      <div className="flex-1">
                        <h3 className="font-medium text-gray-900">{student.name}</h3>
                        <div className="flex items-center space-x-2 text-sm text-gray-500">
                          <span>Grade {student.grade}</span>
                          <span>•</span>
                          <span>{student.studentId}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            
            <div className="flex items-center justify-end space-x-4 p-6 border-t border-gray-200">
              <button
                onClick={() => setShowAddStudent(false)}
                className="btn-secondary"
              >
                Cancel
              </button>
              <button
                onClick={handleAddStudents}
                disabled={selectedStudents.length === 0}
                className="btn-primary"
              >
                <UserPlus className="w-4 h-4 mr-2" />
                Add {selectedStudents.length} Students
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default TeacherClasses; 