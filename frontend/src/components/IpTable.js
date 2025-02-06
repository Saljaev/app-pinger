import React, { useState, useEffect } from 'react';
import { Table, Spin, Alert, Tag } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';

const antIcon = <LoadingOutlined style={{ fontSize: 24 }} spin />;

const IpTable = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const REFRESH_INTERVAL = process.env.REACT_APP_REFRESH_INTERVAL * 1000;
  const API_LOCATION = process.env.REACT_APP_API_LOCATION || "/api";
  const API_URL = `${API_LOCATION}container/getall`;

  const fetchData = async () => {
    try {
      const response = await axios.get(API_URL);
      const formattedData = response.data.map(item => ({
        ip: item.ip_address,
        isReachable: item.is_reachable,
        lastPing: item.last_ping,
      }));
      setData(formattedData);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, REFRESH_INTERVAL);
    return () => clearInterval(interval);
  }, []);

  const columns = [
    {
      title: 'IP Address',
      dataIndex: 'ip',
      key: 'ip',
      sorter: (a, b) => a.ip.localeCompare(b.ip),
    },
    {
      title: 'Reachable',
      dataIndex: 'isReachable',
      key: 'isReachable',
      render: (value) => (
        value ? <Tag color="green">Yes</Tag> : <Tag color="red">No</Tag>
      ),
      filters: [
        { text: 'Yes', value: true },
        { text: 'No', value: false },
      ],
      onFilter: (value, record) => record.isReachable === value,
    },
    {
      title: 'Last Ping',
      dataIndex: 'lastPing',
      key: 'lastPing',
      render: (value) => moment(value).format('YYYY-MM-DD HH:mm:ss'),
      sorter: (a, b) => moment(a.lastPing).unix() - moment(b.lastPing).unix(),
    },
  ];

  if (loading) {
    return <Spin indicator={antIcon} />;
  }

  if (error) {
    return <Alert message={`Error: ${error}`} type="error" />;
  }

  return (
    <div style={{ padding: '20px' }}>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="ip"
        bordered
        pagination={{ pageSize: 10 }}
        scroll={{ x: true }}
      />
    </div>
  );
};

export default IpTable;
