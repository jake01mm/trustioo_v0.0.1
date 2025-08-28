import React from 'react'
import { RouterProvider } from './router'
import Providers from './providers'

function App() {
  return (
    <Providers>
      <RouterProvider />
    </Providers>
  )
}

export default App
