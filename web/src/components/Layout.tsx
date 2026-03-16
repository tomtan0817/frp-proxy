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
      label: '我的域名',
    },
    ...(admin
      ? [
          {
            key: '/admin/users',
            icon: <UserOutlined />,
            label: '用户管理',
          },
          {
            key: '/admin/domains',
            icon: <AppstoreOutlined />,
            label: '域名管理',
          },
          {
            key: '/admin/invites',
            icon: <KeyOutlined />,
            label: '邀请码',
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
            {user?.username}（{user?.role === 'admin' ? '管理员' : '用户'}）
          </Text>
          <Button icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Button>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: '#fff', borderRadius: 8 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}
