import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, InputNumber, message, Tag, Space, Popconfirm } from 'antd';
import { PlusOutlined, CheckOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { getUsers, createUser, updateUser, activateUser, deleteUser } from '../api';

interface User {
  id: number;
  username: string;
  role: string;
  status: string;
  max_domains: number;
  created_at: string;
}

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await getUsers(statusFilter || undefined);
      setUsers(res.data || []);
    } catch {
      message.error('加载用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [statusFilter]);

  const handleCreate = async (values: any) => {
    try {
      await createUser(values);
      message.success('用户创建成功');
      setCreateOpen(false);
      createForm.resetFields();
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || '创建用户失败');
    }
  };

  const handleEdit = async (values: any) => {
    if (!editingUser) return;
    try {
      await updateUser(editingUser.id, values);
      message.success('用户更新成功');
      setEditOpen(false);
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || '更新用户失败');
    }
  };

  const handleActivate = async (id: number) => {
    try {
      await activateUser(id);
      message.success('用户已激活');
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || '激活失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteUser(id);
      message.success('用户已删除');
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const statusColor = (s: string) => {
    switch (s) {
      case 'active': return 'green';
      case 'pending': return 'orange';
      case 'disabled': return 'red';
      default: return 'default';
    }
  };

  const statusLabel = (s: string) => {
    switch (s) {
      case 'active': return '已激活';
      case 'pending': return '待审核';
      case 'disabled': return '已禁用';
      default: return s;
    }
  };

  const roleLabel = (r: string) => r === 'admin' ? '管理员' : '用户';

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (r: string) => <Tag color={r === 'admin' ? 'blue' : 'default'}>{roleLabel(r)}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={statusColor(s)}>{statusLabel(s)}</Tag>,
    },
    { title: '域名上限', dataIndex: 'max_domains', key: 'max_domains' },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => new Date(t).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, r: User) => (
        <Space>
          {r.status === 'pending' && (
            <Button size="small" type="primary" icon={<CheckOutlined />} onClick={() => handleActivate(r.id)}>
              激活
            </Button>
          )}
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingUser(r);
              editForm.setFieldsValue({ role: r.role, status: r.status, max_domains: r.max_domains });
              setEditOpen(true);
            }}
          />
          <Popconfirm title="确定删除此用户？" onConfirm={() => handleDelete(r.id)} okText="确定" cancelText="取消">
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>用户管理</h2>
        <Space>
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: '', label: '全部' },
              { value: 'pending', label: '待审核' },
              { value: 'active', label: '已激活' },
              { value: 'disabled', label: '已禁用' },
            ]}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
            创建用户
          </Button>
        </Space>
      </div>
      <Table dataSource={users} columns={columns} rowKey="id" loading={loading} />

      <Modal title="创建用户" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => createForm.submit()} destroyOnClose okText="确定" cancelText="取消">
        <Form form={createForm} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="角色" initialValue="user">
            <Select options={[{ value: 'user', label: '用户' }, { value: 'admin', label: '管理员' }]} />
          </Form.Item>
          <Form.Item name="max_domains" label="域名上限" initialValue={5}>
            <InputNumber min={1} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="编辑用户" open={editOpen} onCancel={() => setEditOpen(false)} onOk={() => editForm.submit()} destroyOnClose okText="确定" cancelText="取消">
        <Form form={editForm} onFinish={handleEdit} layout="vertical" preserve={false}>
          <Form.Item name="role" label="角色">
            <Select options={[{ value: 'user', label: '用户' }, { value: 'admin', label: '管理员' }]} />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={[{ value: 'active', label: '已激活' }, { value: 'pending', label: '待审核' }, { value: 'disabled', label: '已禁用' }]} />
          </Form.Item>
          <Form.Item name="max_domains" label="域名上限">
            <InputNumber min={1} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
