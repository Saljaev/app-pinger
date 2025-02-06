import React from 'react';
import { ConfigProvider } from 'antd';
import IpTable from './components/IpTable';

function App() {
  return (
    <ConfigProvider theme={{
      token: {
        colorPrimary: '#00b96b',
      },
    }}>
      <div className="App">
        <h1 style={{ textAlign: 'center', padding: '20px' }}>Containers Monitor</h1>
        <IpTable />
      </div>
    </ConfigProvider>
  );
}

export default App;