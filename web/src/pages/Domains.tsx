import { useEffect, useState } from 'react';
import { Table, Button, Modal, Input, message, Tag, Space, Popconfirm, Typography } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined } from '@ant-design/icons';
import { getDomains, createDomain, deleteDomain, getConfig } from '../api';

const { Paragraph } = Typography;

interface Domain {
  id: number;
  subdomain: string;
  token: string;
  status: string;
  created_at: string;
}

export default function Domains() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const [configOpen, setConfigOpen] = useState(false);
  const [subdomain, setSubdomain] = useState('');
  const [newDomain, setNewDomain] = useState<Domain | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [baseDomain, setBaseDomain] = useState('');

  const fetchDomains = async () => {
    setLoading(true);
    try {
      const res = await getDomains();
      setDomains(res.data || []);
    } catch {
      message.error('加载域名失败');
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

  const handleAdd = async () => {
    if (!subdomain.trim()) {
      message.warning('请输入子域名');
      return;
    }
    setSubmitting(true);
    try {
      const res = await createDomain(subdomain.trim());
      message.success('域名创建成功');
      setNewDomain(res.data);
      setAddOpen(false);
      setSubdomain('');
      setConfigOpen(true);
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || '创建域名失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteDomain(id);
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

  const frpcConfig = (d: Domain) =>
    `serverAddr = "frps.${baseDomain || 'example.com'}"
serverPort = 7000
metadatas.token = "${d.token}"

[[proxies]]
name = "${d.subdomain}"
type = "http"
localPort = 3000
subdomain = "${d.subdomain}"`;

  const columns = [
    { title: '子域名', dataIndex: 'subdomain', key: 'subdomain' },
    {
      title: '完整域名',
      key: 'full_domain',
      render: (_: any, r: Domain) => `${r.subdomain}.${baseDomain || 'example.com'}`,
    },
    {
      title: 'Token',
      dataIndex: 'token',
      key: 'token',
      render: (token: string) => (
        <Space>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {token?.substring(0, 16) || ''}...
          </span>
          <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(token)} />
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => (
        <Tag color={s === 'active' ? 'green' : s === 'disabled' ? 'red' : 'default'}>
          {s === 'active' ? '启用' : s === 'disabled' ? '禁用' : s}
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
      render: (_: any, r: Domain) => (
        <Space>
          <Button size="small" onClick={() => { setNewDomain(r); setConfigOpen(true); }}>
            配置
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
        <h2 style={{ margin: 0 }}>我的域名</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setAddOpen(true)}>
          添加域名
        </Button>
      </div>
      <Table dataSource={domains} columns={columns} rowKey="id" loading={loading} />

      <Modal
        title="添加域名"
        open={addOpen}
        onOk={handleAdd}
        onCancel={() => { setAddOpen(false); setSubdomain(''); }}
        confirmLoading={submitting}
        okText="确定"
        cancelText="取消"
      >
        <Input
          placeholder="请输入子域名"
          value={subdomain}
          onChange={(e) => setSubdomain(e.target.value)}
          onPressEnter={handleAdd}
        />
      </Modal>

      <Modal
        title="FRPC 配置"
        open={configOpen}
        onCancel={() => setConfigOpen(false)}
        footer={[
          <Button key="copy" type="primary" onClick={() => newDomain && copyToClipboard(frpcConfig(newDomain))}>
            复制配置
          </Button>,
          <Button key="close" onClick={() => setConfigOpen(false)}>
            关闭
          </Button>,
        ]}
        width={600}
      >
        {newDomain && (
          <Paragraph>
            <pre style={{ background: '#f5f5f5', padding: 16, borderRadius: 8, overflow: 'auto' }}>
              {frpcConfig(newDomain)}
            </pre>
          </Paragraph>
        )}
      </Modal>
    </>
  );
}
