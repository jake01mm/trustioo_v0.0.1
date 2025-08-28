import React from 'react'
import { ConfigProvider, theme } from 'antd'

interface Props {
  children: React.ReactNode
}

function Providers({ children }: Props) {
  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm
      }}
    >
      {children}
    </ConfigProvider>
  )
}

export default Providers
