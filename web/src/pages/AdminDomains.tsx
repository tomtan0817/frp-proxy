import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, message, Tag, Space, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { getAllDomains, adminCreateDomain, adminUpdateDomain, adminDeleteDomain, getConfig } from '../api';

interface DomainRecord {
  id: number;
  subdomain: string;
  user_id: number;
  user?: { username: string };
  token: string;
  status: string;
  created_at: string;
}

export default function AdminDomains() {
  const [domains, setDomains] = useState<DomainRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [form] = Form.useForm();
  const [baseDomain, setBaseDomain] = useState('');

  const fetchDomains = async () => {
    setLoading(true);
    try {
      const res = await getAllDomains();
      setDomains(res.data || []);
    } catch {
      message.error('加载域名列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDomains();
    getConfig().then(res => {
      setBaseDomain(res.data?.base_domain || '');
    }).catch(() => {});
  }, []);

  const handleCreate = async (values: any) => {
    try {
      await adminCreateDomain(values);
      message.success('域名创建成功');
      setCreateOpen(false);
      form.resetFields();
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || '创建域名失败');
    }
  };

  const handleToggleStatus = async (record: DomainRecord) => {
    const newStatus = record.status === 'active' ? 'disabled' : 'active';
    try {
      await adminUpdateDomain(record.id, { status: newStatus });
      message.success(`域名已${newStatus === 'active' ? '启用' : '禁用'}`);
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || '更新失败');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await adminDeleteDomain(id);
      message.success('域名已删除');
      fetchDomains();
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
    { title: '子域名', dataIndex: 'subdomain', key: 'subdomain' },
    {
      title: '完整域名',
      key: 'full_domain',
      render: (_: any, r: DomainRecord) => `${r.subdomain}.${baseDomain || 'example.com'}`,
    },
    {
      title: '所属用户',
      key: 'username',
      render: (_: any, record: any) => record.user?.username || '-',
    },
    {
      title: 'Token',
      dataIndex: 'token',
      key: 'token',
      render: (token: string) => (
        <Button size="small" type="link" onClick={() => copyToClipboard(token)}>
          {token?.substring(0, 16)}... 复制
        </Button>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => (
        <Tag color={s === 'active' ? 'green' : 'red'}>
          {s === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
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
      render: (_: any, r: DomainRecord) => (
        <Space>
          <Button size="small" onClick={() => handleToggleStatus(r)}>
            {r.status === 'active' ? '禁用' : '启用'}
          </Button>
          <Popconfirm title="确定删除此域名？" onConfirm={() => handleDelete(r.id)} okText="确定" cancelText="取消">
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>域名管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          创建域名
        </Button>
      </div>
      <Table dataSource={domains} columns={columns} rowKey="id" loading={loading} />

      <Modal title="创建域名" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => form.submit()} destroyOnClose okText="确定" cancelText="取消">
        <Form form={form} onFinish={handleCreate} layout="vertical" preserve={false}>
          <Form.Item name="user_id" label="用户 ID" rules={[{ required: true, message: '请输入用户 ID' }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="subdomain" label="子域名" rules={[{ required: true, message: '请输入子域名' }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
