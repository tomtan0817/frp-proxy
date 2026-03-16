import { useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, Typography } from 'antd';
import {
  GlobalOutlined,
  UserOutlined,
  KeyOutlined,
  AppstoreOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { getUser, isAdmin, removeToken } from '../auth';

const { Sider, Header, Content } = Layout;
const { Text } = Typography;

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const location = useLocation();
  const user = getUser();
  const admin = isAdmin();

  const handleLogout = () => {
    removeToken();
    navigate('/login');
  };

  const menuItems = [
    {
      key: '/domains',
      icon: <GlobalOutlined />,
      label: 'My Domains',
    },
    ...(admin
      ? [
          {
            key: '/admin/users',
            icon: <UserOutlined />,
            label: 'Users',
          },
          {
            key: '/admin/domains',
            icon: <AppstoreOutlined />,
            label: 'Domains',
          },
          {
            key: '/admin/invites',
            icon: <KeyOutlined />,
            label: 'Invite Codes',
          },
        ]
      : []),
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={0}>
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <Text strong style={{ color: '#fff', fontSize: 18 }}>
            FRP Proxy
          </Text>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            gap: 16,
          }}
        >
          <Text>
            {user?.username} ({user?.role})
          </Text>
          <Button icon={<LogoutOutlined />} onClick={handleLogout}>
            Logout
          </Button>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: '#fff', borderRadius: 8 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}
