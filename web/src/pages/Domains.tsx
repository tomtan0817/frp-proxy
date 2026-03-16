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
      message.error('Failed to load domains');
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
      message.warning('Please enter a subdomain');
      return;
    }
    setSubmitting(true);
    try {
      const res = await createDomain(subdomain.trim());
      message.success('Domain created');
      setNewDomain(res.data);
      setAddOpen(false);
      setSubdomain('');
      setConfigOpen(true);
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to create domain');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteDomain(id);
      message.success('Domain deleted');
      fetchDomains();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to delete');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      message.success('Copied!');
    }).catch(() => {
      message.error('Copy failed');
    });
  };

  const frpcConfig = (d: Domain) =>
    `serverAddr = "${baseDomain || 'example.com'}"
serverPort = 7000
metadatas.token = "${d.token}"

[[proxies]]
name = "web"
type = "http"
localPort = 3000
subdomain = "${d.subdomain}"`;

  const columns = [
    { title: 'Subdomain', dataIndex: 'subdomain', key: 'subdomain' },
    {
      title: 'Full Domain',
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
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => (
        <Tag color={s === 'active' ? 'green' : s === 'disabled' ? 'red' : 'default'}>{s}</Tag>
      ),
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
      render: (_: any, r: Domain) => (
        <Space>
          <Button size="small" onClick={() => { setNewDomain(r); setConfigOpen(true); }}>
            Config
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
        <h2 style={{ margin: 0 }}>My Domains</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setAddOpen(true)}>
          Add Domain
        </Button>
      </div>
      <Table dataSource={domains} columns={columns} rowKey="id" loading={loading} />

      <Modal
        title="Add Domain"
        open={addOpen}
        onOk={handleAdd}
        onCancel={() => { setAddOpen(false); setSubdomain(''); }}
        confirmLoading={submitting}
      >
        <Input
          placeholder="Enter subdomain"
          value={subdomain}
          onChange={(e) => setSubdomain(e.target.value)}
          onPressEnter={handleAdd}
        />
      </Modal>

      <Modal
        title="FRPC Configuration"
        open={configOpen}
        onCancel={() => setConfigOpen(false)}
        footer={[
          <Button key="copy" type="primary" onClick={() => newDomain && copyToClipboard(frpcConfig(newDomain))}>
            Copy Config
          </Button>,
          <Button key="close" onClick={() => setConfigOpen(false)}>
            Close
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
