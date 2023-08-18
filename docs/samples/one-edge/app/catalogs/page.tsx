import React from 'react'
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import CatalogList from '@/components/CatalogList';
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
      spec: catalog.spec.properties['spec'],
    }
  });
  return catalogs;
}
async function CatalogsPage() {
  const [solutionCatalogs, instanceCatalogs] = await Promise.all([getCatalogs('solution'), getCatalogs('instance')]);
  return (
    <div>
      <h1>ABC</h1>
      <h1>Solution Templates</h1>
      <hr/>
      <h1>BBB</h1>       
      <CatalogList catalogs={solutionCatalogs} key="solutions"/>      
      <h1>CCC</h1>       
      <h1>Instance Templates</h1>
      <hr/>
      <CatalogList catalogs={instanceCatalogs} key="instances"/>     
      <h1>DEF</h1>       
    </div>
  );
}

export default CatalogsPage;