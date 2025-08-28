import React from 'react'
import { useRoutes, Navigate } from 'react-router-dom'
import AdminLayout from '@layouts/AdminLayout'

function Dashboard() {
  return <div>Dashboard Page</div>
}

function Login() {
  return <div>Login Page</div>
}

export function RouterProvider() {
  const element = useRoutes([
    {
      path: '/',
      element: <AdminLayout />,
      children: [
        { index: true, element: <Navigate to="/dashboard" replace /> },
        { path: 'dashboard', element: <Dashboard /> }
      ]
    },
    { path: '/login', element: <Login /> },
    { path: '*', element: <div>404 Not Found</div> }
  ])
  return element
}

export default RouterProvider
