import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Readme from '@site/README.md';

export default function Home(): React.JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Self-hosted read-later service with native Kindle delivery">
      <main style={{padding: '2rem 0'}}>
        <div className="container">
          <Readme />
        </div>
      </main>
    </Layout>
  );
}
