import React, { useState } from 'react';
import { 
  TrendingUp, 
  TrendingDown,
  Award, 
  Target, 
  Calendar,
  BookOpen,
  CheckCircle,
  Clock,
  Star,
  ArrowUp,
  ArrowDown,
  Plus,
  Filter,
  BarChart3,
  PieChart,
  Activity,
  Zap,
  Trophy,
  Brain
} from 'lucide-react';

interface Subject {
  id: string;
  name: string;
  currentGrade: number;
  previousGrade: number;
  assignments: number;
  completed: number;
  color: string;
}

interface Goal {
  id: string;
  title: string;
  description: string;
  targetDate: string;
  progress: number;
  category: 'academic' | 'skill' | 'activity';
  priority: 'high' | 'medium' | 'low';
}

interface Achievement {
  id: string;
  title: string;
  description: string;
  date: string;
  type: 'grade' | 'assignment' | 'attendance' | 'improvement';
  icon: string;
}

const StudentProgress: React.FC = () => {
  const [selectedPeriod, setSelectedPeriod] = useState<'week' | 'month' | 'semester'>('month');
  const [showGoalModal, setShowGoalModal] = useState(false);

  // Mock data
  const subjects: Subject[] = [
    {
      id: 'math',
      name: 'Mathematics',
      currentGrade: 87,
      previousGrade: 82,
      assignments: 12,
      completed: 11,
      color: 'blue'
    },
    {
      id: 'science',
      name: 'Science',
      currentGrade: 91,
      previousGrade: 89,
      assignments: 8,
      completed: 8,
      color: 'green'
    },
    {
      id: 'english',
      name: 'English',
      currentGrade: 85,
      previousGrade: 87,
      assignments: 10,
      completed: 9,
      color: 'purple'
    },
    {
      id: 'history',
      name: 'History',
      currentGrade: 92,
      previousGrade: 88,
      assignments: 6,
      completed: 6,
      color: 'orange'
    }
  ];

  const goals: Goal[] = [
    {
      id: 'goal1',
      title: 'Improve Math Grade',
      description: 'Achieve 90% or higher in Mathematics',
      targetDate: '2024-02-15',
      progress: 75,
      category: 'academic',
      priority: 'high'
    },
    {
      id: 'goal2',
      title: 'Complete Science Project',
      description: 'Finish the renewable energy research project',
      targetDate: '2024-01-30',
      progress: 60,
      category: 'academic',
      priority: 'medium'
    },
    {
      id: 'goal3',
      title: 'Reading Challenge',
      description: 'Read 5 books this semester',
      targetDate: '2024-03-01',
      progress: 40,
      category: 'skill',
      priority: 'low'
    }
  ];

  const achievements: Achievement[] = [
    {
      id: 'ach1',
      title: 'Perfect Attendance',
      description: 'No absences for 2 weeks',
      date: '2024-01-15',
      type: 'attendance',
      icon: '🎯'
    },
    {
      id: 'ach2',
      title: 'Top Score',
      description: 'Highest grade in Chemistry quiz',
      date: '2024-01-12',
      type: 'grade',
      icon: '🏆'
    },
    {
      id: 'ach3',
      title: 'Improvement Star',
      description: 'Math grade improved by 10%',
      date: '2024-01-10',
      type: 'improvement',
      icon: '⭐'
    }
  ];

  const overallGPA = subjects.reduce((acc, subject) => acc + subject.currentGrade, 0) / subjects.length;
  const previousGPA = subjects.reduce((acc, subject) => acc + subject.previousGrade, 0) / subjects.length;
  const gpaChange = overallGPA - previousGPA;

  const getColorClasses = (color: string) => {
    const colors = {
      blue: 'bg-blue-500 text-blue-600 border-blue-200 bg-blue-50',
      green: 'bg-green-500 text-green-600 border-green-200 bg-green-50',
      purple: 'bg-purple-500 text-purple-600 border-purple-200 bg-purple-50',
      orange: 'bg-orange-500 text-orange-600 border-orange-200 bg-orange-50'
    };
    return colors[color as keyof typeof colors] || colors.blue;
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'text-red-600 bg-red-50 border-red-200';
      case 'medium': return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'low': return 'text-green-600 bg-green-50 border-green-200';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  return (
    <div className="space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold text-gray-900">
            My Progress 📊
          </h1>
          <p className="text-gray-600 text-lg">
            Track your academic journey and achievements
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <select
            value={selectedPeriod}
            onChange={(e) => setSelectedPeriod(e.target.value as any)}
            className="input w-auto"
          >
            <option value="week">This Week</option>
            <option value="month">This Month</option>
            <option value="semester">This Semester</option>
          </select>
          <button
            onClick={() => setShowGoalModal(true)}
            className="btn-primary"
          >
            <Plus className="w-4 h-4 mr-2" />
            Set Goal
          </button>
        </div>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="card p-6 bg-gradient-to-br from-blue-50 to-blue-100 border-blue-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-blue-500 rounded-2xl flex items-center justify-center">
              <Award className="w-6 h-6 text-white" />
            </div>
            <div className={`flex items-center space-x-1 text-sm font-medium ${
              gpaChange >= 0 ? 'text-green-600' : 'text-red-600'
            }`}>
              {gpaChange >= 0 ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />}
              <span>{Math.abs(gpaChange).toFixed(1)}</span>
            </div>
          </div>
          <div className="space-y-1">
            <h3 className="text-2xl font-bold text-gray-900">{overallGPA.toFixed(1)}</h3>
            <p className="text-sm font-medium text-gray-600">Overall GPA</p>
            <p className="text-xs text-gray-500">Out of 100</p>
          </div>
        </div>

        <div className="card p-6 bg-gradient-to-br from-green-50 to-green-100 border-green-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-green-500 rounded-2xl flex items-center justify-center">
              <CheckCircle className="w-6 h-6 text-white" />
            </div>
            <div className="text-sm font-medium text-green-600">
              <span>96%</span>
            </div>
          </div>
          <div className="space-y-1">
            <h3 className="text-2xl font-bold text-gray-900">
              {subjects.reduce((acc, s) => acc + s.completed, 0)}
            </h3>
            <p className="text-sm font-medium text-gray-600">Completed</p>
            <p className="text-xs text-gray-500">
              of {subjects.reduce((acc, s) => acc + s.assignments, 0)} assignments
            </p>
          </div>
        </div>

        <div className="card p-6 bg-gradient-to-br from-purple-50 to-purple-100 border-purple-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-purple-500 rounded-2xl flex items-center justify-center">
              <Target className="w-6 h-6 text-white" />
            </div>
            <div className="text-sm font-medium text-purple-600">
              <span>3 active</span>
            </div>
          </div>
          <div className="space-y-1">
            <h3 className="text-2xl font-bold text-gray-900">
              {Math.round(goals.reduce((acc, g) => acc + g.progress, 0) / goals.length)}%
            </h3>
            <p className="text-sm font-medium text-gray-600">Goals Progress</p>
            <p className="text-xs text-gray-500">Average completion</p>
          </div>
        </div>

        <div className="card p-6 bg-gradient-to-br from-orange-50 to-orange-100 border-orange-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-orange-500 rounded-2xl flex items-center justify-center">
              <Trophy className="w-6 h-6 text-white" />
            </div>
            <div className="text-sm font-medium text-orange-600">
              <span>This month</span>
            </div>
          </div>
          <div className="space-y-1">
            <h3 className="text-2xl font-bold text-gray-900">{achievements.length}</h3>
            <p className="text-sm font-medium text-gray-600">Achievements</p>
            <p className="text-xs text-gray-500">Earned recently</p>
          </div>
        </div>
      </div>

      {/* Subject Performance */}
      <div className="card">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              Subject Performance
            </h2>
            <div className="flex items-center space-x-2">
              <BarChart3 className="w-5 h-5 text-gray-400" />
              <span className="text-sm text-gray-500">Current Semester</span>
            </div>
          </div>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {subjects.map((subject) => {
              const colors = getColorClasses(subject.color).split(' ');
              const bgColor = colors[0];
              const textColor = colors[1];
              const borderColor = colors[2];
              const lightBg = colors[3];
              
              return (
                <div
                  key={subject.id}
                  className={`p-6 rounded-2xl border-2 ${borderColor} ${lightBg} hover:shadow-lg transition-all`}
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <div className={`w-10 h-10 ${bgColor} rounded-xl flex items-center justify-center`}>
                        <BookOpen className="w-5 h-5 text-white" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-gray-900">{subject.name}</h3>
                        <p className="text-sm text-gray-500">
                          {subject.completed}/{subject.assignments} assignments
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-gray-900">
                        {subject.currentGrade}%
                      </div>
                      <div className={`flex items-center space-x-1 text-sm font-medium ${
                        subject.currentGrade >= subject.previousGrade ? 'text-green-600' : 'text-red-600'
                      }`}>
                        {subject.currentGrade >= subject.previousGrade ? (
                          <ArrowUp className="w-4 h-4" />
                        ) : (
                          <ArrowDown className="w-4 h-4" />
                        )}
                        <span>{Math.abs(subject.currentGrade - subject.previousGrade)}</span>
                      </div>
                    </div>
                  </div>
                  
                  {/* Progress Bar */}
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-600">Progress</span>
                      <span className={textColor}>
                        {Math.round((subject.completed / subject.assignments) * 100)}%
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className={`${bgColor} h-2 rounded-full transition-all duration-300`}
                        style={{ width: `${(subject.completed / subject.assignments) * 100}%` }}
                      ></div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Goals and Achievements */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Goals */}
        <div className="card">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900">
                Learning Goals
              </h2>
              <button
                onClick={() => setShowGoalModal(true)}
                className="btn-ghost text-sm"
              >
                <Plus className="w-4 h-4 mr-1" />
                Add Goal
              </button>
            </div>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {goals.map((goal) => (
                <div
                  key={goal.id}
                  className="p-4 rounded-xl border border-gray-200 hover:border-gray-300 transition-colors"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-1">
                        <h3 className="font-semibold text-gray-900">{goal.title}</h3>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getPriorityColor(goal.priority)}`}>
                          {goal.priority}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mb-2">{goal.description}</p>
                      <div className="flex items-center space-x-4 text-xs text-gray-500">
                        <div className="flex items-center space-x-1">
                          <Calendar className="w-3 h-3" />
                          <span>Due {new Date(goal.targetDate).toLocaleDateString()}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <Target className="w-3 h-3" />
                          <span>{goal.category}</span>
                        </div>
                      </div>
                    </div>
                    <div className="text-right ml-4">
                      <div className="text-lg font-bold text-blue-600">{goal.progress}%</div>
                    </div>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${goal.progress}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Achievements */}
        <div className="card">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900">
                Recent Achievements
              </h2>
              <div className="flex items-center space-x-2">
                <Trophy className="w-5 h-5 text-yellow-500" />
                <span className="text-sm text-gray-500">{achievements.length} total</span>
              </div>
            </div>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {achievements.map((achievement) => (
                <div
                  key={achievement.id}
                  className="p-4 rounded-xl bg-gradient-to-r from-yellow-50 to-orange-50 border border-yellow-200 hover:shadow-md transition-all"
                >
                  <div className="flex items-start space-x-4">
                    <div className="text-2xl">{achievement.icon}</div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between mb-1">
                        <h3 className="font-semibold text-gray-900">{achievement.title}</h3>
                        <span className="text-xs text-gray-500">
                          {new Date(achievement.date).toLocaleDateString()}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600">{achievement.description}</p>
                      <div className="mt-2">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                          achievement.type === 'grade' ? 'bg-blue-100 text-blue-800' :
                          achievement.type === 'assignment' ? 'bg-green-100 text-green-800' :
                          achievement.type === 'attendance' ? 'bg-purple-100 text-purple-800' :
                          'bg-orange-100 text-orange-800'
                        }`}>
                          {achievement.type}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Study Insights */}
      <div className="card">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              Study Insights
            </h2>
            <div className="flex items-center space-x-2">
              <Brain className="w-5 h-5 text-purple-500" />
              <span className="text-sm text-gray-500">AI-Powered</span>
            </div>
          </div>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center p-6 bg-gradient-to-br from-blue-50 to-blue-100 rounded-2xl">
              <div className="w-12 h-12 bg-blue-500 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <TrendingUp className="w-6 h-6 text-white" />
              </div>
              <h3 className="font-semibold text-gray-900 mb-2">Strongest Subject</h3>
              <p className="text-2xl font-bold text-blue-600">History</p>
              <p className="text-sm text-gray-600">92% average</p>
            </div>
            
            <div className="text-center p-6 bg-gradient-to-br from-orange-50 to-orange-100 rounded-2xl">
              <div className="w-12 h-12 bg-orange-500 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <Target className="w-6 h-6 text-white" />
              </div>
              <h3 className="font-semibold text-gray-900 mb-2">Focus Area</h3>
              <p className="text-2xl font-bold text-orange-600">Mathematics</p>
              <p className="text-sm text-gray-600">Needs attention</p>
            </div>
            
            <div className="text-center p-6 bg-gradient-to-br from-green-50 to-green-100 rounded-2xl">
              <div className="w-12 h-12 bg-green-500 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <h3 className="font-semibold text-gray-900 mb-2">Study Streak</h3>
              <p className="text-2xl font-bold text-green-600">12 days</p>
              <p className="text-sm text-gray-600">Keep it up!</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StudentProgress; 