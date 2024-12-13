import React from 'react'
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import CatalogLists from '@/components/catalogs/CatalogLists';
import {CatalogState, User} from '../types';
const getCatalogs = async (type: string) => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}catalogs/registry`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();
  //map each element and do some transformation  
  
  console.log('Full data JSON:', JSON.stringify(data, null, 2));

  const catalogs = data
  .filter((catalog: CatalogState) => catalog.spec.catalogType === type);
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