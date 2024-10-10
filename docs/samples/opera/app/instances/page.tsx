import React from 'react'
import { getServerSession } from 'next-auth';
import MultiView from '@/components/MultiView';
import { options } from '../api/auth/[...nextauth]/options';
import {InstanceState, User} from '../types';
const getInstances = async () => {
  const session = await getServerSession(options);    
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}instances`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }
  });
  const data = await res.json();  
  return data;
}
async function InstancessPage() {
  const instances = await getInstances();
  const params = {
    type: 'instances',
    menuItems: [
      {
        name: 'Add Solution',
        href: '/solutions/add',                
      }
    ],
    views: ['cards', 'table'],
    items: instances,
    columns: []
  }
  return (
    <div>
        <MultiView params={params}  />
    </div>
);
}

export default InstancessPage;