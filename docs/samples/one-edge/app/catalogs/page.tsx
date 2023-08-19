import React from 'react'
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import CatalogLists from '@/components/CatalogLists';
const getCatalogs = async (type: string) => {
  const session = await getServerSession(options);  
  console.log(session?.user?.accessToken);
  const res = await fetch('http://localhost:8082/v1alpha2/catalogs/registry', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${session?.user?.accessToken}`,
    }
  });
  const data = await res.json();
  //map each element and do some transformation
  const catalogs = data
  .filter((catalog: any) => catalog.spec.type === type)
  .map((catalog: any) => {
    return {
      id: catalog.id,
      name: catalog.spec.name,
      type: catalog.spec.type,
      properties: catalog.spec.properties,
    }
  });
  return catalogs;
}
async function CatalogsPage() {
  const [solutionCatalogs, instanceCatalogs, configCatalogs] = await Promise.all([getCatalogs('solution'), getCatalogs('instance'), getCatalogs('config')]);
  return (
    <div className='cards_view'>
      <CatalogLists groups={[
        {
          catalogs: solutionCatalogs,
          title: 'Solution Templates',
          type: 'solution'
        },
        {
          catalogs: instanceCatalogs,
          title: 'Instance Templates',
          type: 'instance'
        },
        {
          catalogs: configCatalogs,
          title: 'Standard Configurations',
          type: 'config'
        }            
      ]}/>
    </div>
  );
}

export default CatalogsPage;