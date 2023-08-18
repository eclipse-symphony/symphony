import React from 'react'


const getSites = async () => {
  const res = await fetch('https://api.github.com/users');
  const data = await res.json();
  return data;
}

async function SitesPage() {
  const sites = await getSites();  
  return (
    <div>
      {sites.map((site: any) => {
        return (
          <div key={site.id}>
            <h1>{site.login}</h1>
            <p>{site.url}</p>
          </div>
        );
      })}
    </div>
  );
}

export default SitesPage;