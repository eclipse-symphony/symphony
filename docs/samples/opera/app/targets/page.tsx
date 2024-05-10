import React from 'react'
import { getServerSession } from 'next-auth';
import MultiView from '@/components/MultiView';
import { options } from '../api/auth/[...nextauth]/options';
import {TargetState, User} from '../types';
const getTargets = async () => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}targets/registry`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();
  return data;
}
async function TargetsPage() {
  const targets = await getTargets();
  const params = {
    type: 'targets',
    menuItems: [
      {
        name: 'Add Solution',
        href: '/solutions/add',                
      }
    ],
    views: ['cards', 'table'],
    items: targets,
    columns: []
  }
  return (
    <div>
        <MultiView params={params}  />
    </div>
);
}

export default TargetsPage;