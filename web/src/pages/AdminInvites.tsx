import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, InputNumber, message, Space, Popconfirm } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined } from '@ant-design/icons';
import { getInviteCodes, createInviteCode, deleteInviteCode } from '../api';

interface InviteCode {
  id: number;
  code: string;
  max_uses: number;
  used_count: number;
  expires_at: string;
  created_at: string;
}

export default function AdminInvites() {
  const [codes, setCodes] = useState<InviteCode[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [form] = Form.useForm();

  const fetchCodes = async () => {
    setLoading(true);
    try {
      const res = await getInviteCodes();
      setCodes(res.data || []);
    } catch {
      message.error('Failed to load invite codes');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCodes();
  }, []);

  const handleCreate = async (values: any) => {
    try {
      await createInviteCode(values);
      message.success('Invite code created');
      setCreateOpen(false);
      form.resetFields();
      fetchCodes();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to create invite code');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteInviteCode(id);
      message.success('Invite code deleted');
      fetchCodes();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to delete');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => message.success('Copied'));
  };

  const columns = [
    {
      title: 'Code',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => (
        <Space>
          <span style={{ fontFamily: 'monospace' }}>{code}</span>
          <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(code)} />
        </Space>
      ),
    },
    { title: 'Max Uses', dataIndex: 'max_uses', key: 'max_uses' },
    { title: 'Used', dataIndex: 'used_count', key: 'used_count' },
    {
      title: 'Expires At',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (t: string) => (t ? new Date(t).toLocaleString() : 'Never'),
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
      render: (_: any, r: InviteCode) => (
        <Space>
          <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(r.code)}>
            Copy
          </Button>
          <Popconfirm title="Delete this invite code?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>Invite Codes</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          Generate Code
        </Button>
      </div>
      <Table dataSource={codes} columns={columns} rowKey="id" loading={loading} />

      <Modal title="Generate Invite Code" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => form.submit()} destroyOnClose>
        <Form form={form} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="max_uses" label="Max Uses" initialValue={1}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="expires_in_hours" label="Expires In (hours)" initialValue={72}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
