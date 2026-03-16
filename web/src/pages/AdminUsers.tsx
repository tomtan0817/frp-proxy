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
      message.error('Failed to load users');
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
      message.success('User created');
      setCreateOpen(false);
      createForm.resetFields();
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to create user');
    }
  };

  const handleEdit = async (values: any) => {
    if (!editingUser) return;
    try {
      await updateUser(editingUser.id, values);
      message.success('User updated');
      setEditOpen(false);
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to update user');
    }
  };

  const handleActivate = async (id: number) => {
    try {
      await activateUser(id);
      message.success('User activated');
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to activate');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteUser(id);
      message.success('User deleted');
      fetchUsers();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to delete');
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

  const columns = [
    { title: 'Username', dataIndex: 'username', key: 'username' },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (r: string) => <Tag color={r === 'admin' ? 'blue' : 'default'}>{r}</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={statusColor(s)}>{s}</Tag>,
    },
    { title: 'Max Domains', dataIndex: 'max_domains', key: 'max_domains' },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => new Date(t).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, r: User) => (
        <Space>
          {r.status === 'pending' && (
            <Button size="small" type="primary" icon={<CheckOutlined />} onClick={() => handleActivate(r.id)}>
              Activate
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
          <Popconfirm title="Delete this user?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>Users</h2>
        <Space>
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: '', label: 'All' },
              { value: 'pending', label: 'Pending' },
              { value: 'active', label: 'Active' },
              { value: 'disabled', label: 'Disabled' },
            ]}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
            Create User
          </Button>
        </Space>
      </div>
      <Table dataSource={users} columns={columns} rowKey="id" loading={loading} />

      <Modal title="Create User" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => createForm.submit()} destroyOnClose>
        <Form form={createForm} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="username" label="Username" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="Password" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="Role" initialValue="user">
            <Select options={[{ value: 'user', label: 'User' }, { value: 'admin', label: 'Admin' }]} />
          </Form.Item>
          <Form.Item name="max_domains" label="Max Domains" initialValue={5}>
            <InputNumber min={1} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="Edit User" open={editOpen} onCancel={() => setEditOpen(false)} onOk={() => editForm.submit()} destroyOnClose>
        <Form form={editForm} onFinish={handleEdit} layout="vertical" preserve={false}>
          <Form.Item name="role" label="Role">
            <Select options={[{ value: 'user', label: 'User' }, { value: 'admin', label: 'Admin' }]} />
          </Form.Item>
          <Form.Item name="status" label="Status">
            <Select options={[{ value: 'active', label: 'Active' }, { value: 'pending', label: 'Pending' }, { value: 'disabled', label: 'Disabled' }]} />
          </Form.Item>
          <Form.Item name="max_domains" label="Max Domains">
            <InputNumber min={1} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
