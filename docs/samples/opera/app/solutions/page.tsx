import React from 'react'
import { getServerSession } from 'next-auth';
import MultiView from '@/components/MultiView';
import { options } from '../api/auth/[...nextauth]/options';
import {SolutionState, User} from '../types';
const getSolutions = async () => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}solutions`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();
  return data;
}
async function SolutionsPage() {
  const solutions = await getSolutions();
  const params = {
    type: 'solutions',
    menuItems: [
      {
        name: 'Add Solution',
        href: '/solutions/add',                
      }
    ],
    views: ['cards', 'table'],
    items: solutions,
    columns: []
  }
  return (
    <div>
        <MultiView params={params}  />
    </div>
);
}

export default SolutionsPage;