package service

/*

📁 alumni-management-system/
├── 📁 src/
│   ├── 📁 components/
│   │   ├── ProtectedRoute.tsx      # Route protection component
│   │   └── Sidebar.tsx             # Navigation sidebar component
│   │
│   ├── 📁 data/                    # Mock data storage
│   │   ├── eventData.ts            # Event mock data
│   │   ├── filterConfig.ts         # Filter configuration
│   │   ├── galleryData.ts          # Gallery mock data
│   │   ├── sampleData.ts           # Alumni, contacts, news, resources data
│   │   └── schoolsData.ts          # Schools and departments data
│   │
│   ├── 📁 hooks/                   # Custom React hooks
│   │   ├── useAlumniData.ts        # Alumni data fetching
│   │   ├── useEventData.ts         # Event data fetching
│   │   ├── useFilteredAlumni.ts    # Filtered alumni data
│   │   └── useGalleryData.ts       # Gallery data fetching
│   │
│   ├── 📁 layouts/                 # Layout components
│   │   ├── AdminLayout.tsx         # Admin panel layout
│   │   ├── MainLayout.tsx          # Main application layout
│   │   └── PublicLayout.tsx        # Public pages layout
│   │
│   ├── 📁 pages/                   # Page components
│   │   ├── 📁 admin/
│   │   │   └── Dashboard.tsx       # Admin dashboard
│   │   ├── 📁 auth/
│   │   │   └── Login.tsx          # Authentication page
│   │   ├── 📁 public/
│   │   │   └── AlumniPortal.tsx   # Public alumni directory
│   │   └── Home.tsx               # Homepage
│   │
│   ├── 📁 store/                   # Redux store
│   │   ├── 📁 slices/
│   │   │   ├── alumniSlice.ts     # Alumni state management
│   │   │   └── authSlice.ts       # Authentication state
│   │   └── index.ts               # Store configuration
│   │
│   ├── 📁 types/                   # TypeScript types
│   │   └── index.ts               # Type definitions
│   │
│   ├── App.tsx                     # Main application component
│   ├── main.tsx                    # Application entry point
│   ├── index.css                   # Global styles
│   └── vite-env.d.ts              # Vite type declarations
│
├── 📁 public/                      # Static assets
├── .eslintrc.json                  # ESLint configuration
├── index.html                      # HTML entry point
├── package.json                    # Project dependencies
├── postcss.config.js              # PostCSS configuration
├── tailwind.config.js             # Tailwind CSS configuration
├── tsconfig.json                  # TypeScript configuration
└── vite.config.ts                 # Vite configuration


*/
