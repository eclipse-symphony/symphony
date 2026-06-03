import React from 'react'
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import CatalogVersionLists from '@/components/catalogversions/CatalogVersionLists';
import {CatalogVersionState, User} from '../types';
const getCatalogVersions = async (type: string) => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}catalogversions/registry`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();
  //map each element and do some transformation  
  
  console.log('Full data JSON:', JSON.stringify(data, null, 2));

  const catalogversions = data
  .filter((catalogversion: CatalogVersionState) => catalogversion.spec.catalogType === type);
  return catalogversions;
}
async function CatalogVersionsPage() {
  const [solutionversionCatalogVersions, instanceCatalogVersions, configCatalogVersions] = await Promise.all([getCatalogVersions('solutionversion'), getCatalogVersions('instance'), getCatalogVersions('config')]);
  return (
    <div className='cards_view'>
      <CatalogVersionLists groups={[
        {
          catalogversions: solutionversionCatalogVersions,
          title: 'SolutionVersion Templates',
          type: 'solutionversion'
        },
        {
          catalogversions: instanceCatalogVersions,
          title: 'Instance Templates',
          type: 'instance'
        },
        {
          catalogversions: configCatalogVersions,
          title: 'Standard Configurations',
          type: 'config'
        }            
      ]}/>
    </div>
  );
}

export default CatalogVersionsPage;