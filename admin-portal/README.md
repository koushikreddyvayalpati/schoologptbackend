# SchoolGPT Admin Portal

A modern, high-performance React.js admin portal for SchoolGPT - an AI-powered school management system.

## ✨ Features

- 🤖 **AI-Powered School Setup** - Natural language school configuration
- 🎨 **Modern UI/UX** - Clean, professional design with Tailwind CSS
- ⚡ **Performance Optimized** - Lazy loading, code splitting, minimal memory usage
- 📱 **Responsive Design** - Works perfectly on all devices
- 🔐 **Secure Authentication** - JWT-based auth with role management
- 🏢 **Multi-Tenant Architecture** - Manage multiple schools in isolation
- 📊 **Real-time Dashboard** - Live metrics and system health monitoring
- 🎯 **Accessibility** - WCAG compliant with keyboard navigation

## 🚀 Quick Start

### Prerequisites

- Node.js 16+ 
- npm or yarn
- SchoolGPT backend API running

### Installation

1. **Clone and setup**
```bash
cd admin-portal
npm install
```

2. **Environment configuration**
```bash
cp .env.example .env
# Edit .env with your API endpoints
```

3. **Start development server**
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

## 🏗️ Architecture

### Technology Stack

- **React 18** - Latest React with concurrent features
- **TypeScript** - Type safety and better DX
- **Vite** - Lightning fast build tool
- **Tailwind CSS** - Utility-first CSS framework
- **Zustand** - Lightweight state management
- **React Query** - Server state management
- **React Router 6** - Client-side routing
- **Lucide React** - Beautiful icons

### Performance Optimizations

- **Code Splitting** - Automatic route-based splitting
- **Lazy Loading** - Components loaded on demand
- **Tree Shaking** - Unused code elimination
- **Image Optimization** - Responsive images with lazy loading
- **Bundle Analysis** - Optimized chunk sizes
- **Memory Management** - Efficient state and component lifecycle

### Folder Structure

```
src/
├── components/          # Reusable UI components
│   ├── ui/             # Basic UI elements
│   ├── Layout.tsx      # Main layout wrapper
│   ├── Header.tsx      # App header
│   └── Sidebar.tsx     # Navigation sidebar
├── pages/              # Route components
│   ├── Dashboard.tsx   # Main dashboard
│   ├── Schools.tsx     # Schools management
│   ├── SchoolSetup.tsx # AI school setup
│   └── Login.tsx       # Authentication
├── stores/             # State management
│   └── authStore.ts    # Authentication state
├── lib/                # Utilities and helpers
│   └── utils.ts        # Common utilities
└── App.tsx             # Root component
```

## 🎨 Design System

### Color Palette

- **Primary**: Blue (#0ea5e9) - Main brand color
- **Success**: Green (#22c55e) - Success states
- **Warning**: Yellow (#f59e0b) - Warning states  
- **Danger**: Red (#ef4444) - Error states
- **Gray**: Neutral grays for text and backgrounds

### Typography

- **Font**: Inter - Modern, readable sans-serif
- **Mono**: JetBrains Mono - Code and technical content

### Components

All components follow consistent patterns:
- Semantic HTML structure
- Accessible ARIA labels
- Keyboard navigation support
- Responsive design
- Loading and error states

## 🔧 Development

### Available Scripts

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run preview      # Preview production build
npm run lint         # Run ESLint
npm run type-check   # TypeScript type checking
```

### Code Quality

- **ESLint** - Code linting with React rules
- **TypeScript** - Static type checking
- **Prettier** - Code formatting (via editor)

### State Management

Using Zustand for lightweight, performant state management:

```typescript
// Example store
const useStore = create<State>((set) => ({
  data: [],
  loading: false,
  fetchData: async () => {
    set({ loading: true })
    // API call
    set({ data: result, loading: false })
  }
}))
```

## 🎯 Key Features

### AI School Setup

Interactive chat interface for school configuration:
- Natural language processing
- Intelligent field extraction
- Real-time configuration preview
- One-click school creation

### Dashboard

Comprehensive overview with:
- Key metrics and statistics
- Recent activity feed
- System health monitoring
- Quick action shortcuts

### School Management

Advanced school administration:
- Search and filtering
- Bulk operations
- Status management
- Configuration editing

## 🔐 Security

- JWT token management
- Secure API communication
- Role-based access control
- XSS protection
- CSRF protection

## 🌐 API Integration

The portal integrates with SchoolGPT backend API:

```typescript
// Example API call
const response = await fetch('/api/v1/schools', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
```

## 📱 Responsive Design

Breakpoints:
- **Mobile**: 0-640px
- **Tablet**: 641-1024px  
- **Desktop**: 1025px+

## ♿ Accessibility

- WCAG 2.1 AA compliance
- Screen reader support
- Keyboard navigation
- High contrast mode
- Reduced motion support

## 🚀 Deployment

### Production Build

```bash
npm run build
```

Generates optimized static files in `dist/` directory.

### Environment Variables

```bash
VITE_API_URL=https://api.schoolgpt.com
VITE_APP_NAME=SchoolGPT Admin Portal
```

### Docker Support

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "run", "preview"]
```

## 📈 Performance Metrics

Target performance benchmarks:
- **First Contentful Paint**: < 1.5s
- **Largest Contentful Paint**: < 2.5s
- **Cumulative Layout Shift**: < 0.1
- **First Input Delay**: < 100ms

## 🤝 Contributing

1. Follow the established code patterns
2. Write TypeScript with proper types
3. Include tests for new features
4. Follow accessibility guidelines
5. Optimize for performance

## 📄 License

Private - SchoolGPT Project

---

Built with ❤️ for modern school management 