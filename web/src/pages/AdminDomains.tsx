import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, message, Tag, Space, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { getAllDomains, adminCreateDomain, adminUpdateDomain, adminDeleteDomain } from '../api';

interface DomainRecord {
  id: number;
  subdomain: string;
  domain: string;
  user_id: number;
  username: string;
  token: string;
  status: string;
  created_at: string;
}

export default function AdminDomains() {
  const [domains, setDomains] = useState<DomainRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchDomains = async () => {
    setLoading(true);
    try {
      const res = await getAllDomains();
      setDomains(res.data || []);
    } catch {
      message.error('Failed to load domains');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDomains();
  }, []);

  const handleCreate = async (values: any) => {
    try {
      await adminCreateDomain(values);
      message.success('Domain created');
      setCreateOpen(false);
      form.resetFields();
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to create domain');
    }
  };

  const handleToggleStatus = async (record: DomainRecord) => {
    const newStatus = record.status === 'active' ? 'disabled' : 'active';
    try {
      await adminUpdateDomain(record.id, { status: newStatus });
      message.success(`Domain ${newStatus}`);
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to update');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await adminDeleteDomain(id);
      message.success('Domain deleted');
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to delete');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => message.success('Copied'));
  };

  const columns = [
    { title: 'Subdomain', dataIndex: 'subdomain', key: 'subdomain' },
    { title: 'User', dataIndex: 'username', key: 'username' },
    {
      title: 'Token',
      dataIndex: 'token',
      key: 'token',
      render: (token: string) => (
        <Button size="small" type="link" onClick={() => copyToClipboard(token)}>
          {token?.substring(0, 16)}... Copy
        </Button>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={s === 'active' ? 'green' : 'red'}>{s}</Tag>,
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => new Date(t).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, r: DomainRecord) => (
        <Space>
          <Button size="small" onClick={() => handleToggleStatus(r)}>
            {r.status === 'active' ? 'Disable' : 'Enable'}
          </Button>
          <Popconfirm title="Delete this domain?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>All Domains</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          Create Domain
        </Button>
      </div>
      <Table dataSource={domains} columns={columns} rowKey="id" loading={loading} />

      <Modal title="Create Domain" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => form.submit()} destroyOnClose>
        <Form form={form} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="user_id" label="User ID" rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="subdomain" label="Subdomain" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
