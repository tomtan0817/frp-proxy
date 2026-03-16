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
      message.error('加载邀请码失败');
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
      message.success('邀请码创建成功');
      setCreateOpen(false);
      form.resetFields();
      fetchCodes();
    } catch (err: any) {
      message.error(err.response?.data?.error || '创建邀请码失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteInviteCode(id);
      message.success('邀请码已删除');
      fetchCodes();
    } catch (err: any) {
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      message.success('已复制');
    }).catch(() => {
      message.error('复制失败');
    });
  };

  const columns = [
    {
      title: '邀请码',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => (
        <Space>
          <span style={{ fontFamily: 'monospace' }}>{code}</span>
          <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(code)} />
        </Space>
      ),
    },
    { title: '最大使用次数', dataIndex: 'max_uses', key: 'max_uses' },
    { title: '已使用', dataIndex: 'used_count', key: 'used_count' },
    {
      title: '过期时间',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (t: string) => (t ? new Date(t).toLocaleString() : '永不过期'),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t: string) => new Date(t).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, r: InviteCode) => (
        <Space>
          <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(r.code)}>
            复制
          </Button>
          <Popconfirm title="确定删除此邀请码？" onConfirm={() => handleDelete(r.id)} okText="确定" cancelText="取消">
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>邀请码管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          生成邀请码
        </Button>
      </div>
      <Table dataSource={codes} columns={columns} rowKey="id" loading={loading} />

      <Modal title="生成邀请码" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => form.submit()} destroyOnClose okText="确定" cancelText="取消">
        <Form form={form} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="max_uses" label="最大使用次数" initialValue={1}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="expires_in_hours" label="过期时间（小时）" initialValue={72}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
