import React from 'react'
import { Layout, Menu } from 'antd'
import { Link, Outlet, useLocation } from 'react-router-dom'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: 'dashboard', label: <Link to="/dashboard">仪表盘</Link> }
]

function AdminLayout() {
  const location = useLocation()
  const selectedKeys = [location.pathname.replace('/', '') || 'dashboard']

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible>
        <div style={{ height: 48, margin: 16, background: 'rgba(255,255,255,0.3)' }} />
        <Menu theme="dark" mode="inline" items={menuItems} selectedKeys={selectedKeys} />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: 0 }} />
        <Content style={{ margin: '16px' }}>
          <div style={{ padding: 24, background: '#fff', minHeight: 360 }}>
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}

export default AdminLayout
