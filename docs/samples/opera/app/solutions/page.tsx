import React from 'react'
import { getServerSession } from 'next-auth';
import MultiView from '@/components/MultiView';
import { options } from '../api/auth/[...nextauth]/options';
import {SolutionVersionState, User} from '../types';
const getSolutionVersions = async () => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}solutionversions`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();
  return data;
}
async function SolutionVersionsPage() {
  const solutionversions = await getSolutionVersions();
  const params = {
    type: 'solutionversions',
    menuItems: [
      {
        name: 'Add SolutionVersion',
        href: '/solutionversions/add',                
      }
    ],
    views: ['cards', 'table'],
    items: solutionversions,
    columns: []
  }
  return (
    <div>
        <MultiView params={params}  />
    </div>
);
}

export default SolutionVersionsPage;