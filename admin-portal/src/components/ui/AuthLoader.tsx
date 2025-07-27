import React from 'react';
import { GraduationCap, Loader2 } from 'lucide-react';

const AuthLoader: React.FC = () => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-secondary-50 flex items-center justify-center">
      <div className="text-center">
        {/* Logo */}
        <div className="flex justify-center items-center mb-6">
          <div className="bg-primary-600 p-4 rounded-2xl shadow-lg">
            <GraduationCap className="h-12 w-12 text-white" />
          </div>
        </div>
        
        {/* Loading Animation */}
        <div className="mb-4">
          <Loader2 className="h-8 w-8 text-primary-600 animate-spin mx-auto" />
        </div>
        
        {/* Text */}
        <h2 className="text-xl font-semibold text-gray-900 mb-2">
          Loading SchoolGPT
        </h2>
        <p className="text-gray-600 max-w-sm mx-auto">
          Initializing your educational experience...
        </p>
        
        {/* Progress Dots */}
        <div className="flex justify-center space-x-2 mt-6">
          <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
          <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
          <div className="w-2 h-2 bg-primary-600 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
        </div>
      </div>
    </div>
  );
};

export default AuthLoader; 