import React from 'react'
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import MultiView from '@/components/MultiView';
import { User } from '../types';

const getSites = async () => {
  const session = await getServerSession(options);  
  const symphonyApi = process.env.SYMPHONY_API;
  const userObj: User | undefined = session?.user?? undefined;
  const res = await fetch( `${symphonyApi}federation/registry`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${userObj?.accessToken}`,
    }    
  });
  const data = await res.json();
  //map each element and do some transformation
  const sites = data.map((site: any) => {
    return {
      id: site.id,
      phone: site.spec.properties?.phone ?? '',
      address: site.spec.properties?.address ?? '',
      city: site.spec.properties?.city ?? '',
      state: site.spec.properties?.state ?? '',
      zip: site.spec.properties?.zip ?? '',
      country: site.spec.properties?.country ?? '',
      version: site.spec.properties?.version ?? '',
      name: site.spec.properties?.name ?? (site.spec.properties?.id ?? ''),
      self: site.spec.isSelf ?? false,
      lastReported: site.status?.lastReported ? new Date(site.status.lastReported) : null,
      lat: site.spec.properties?.lat ? parseFloat(site.spec.properties.lat)  : 0,
      lng: site.spec.properties?.lng ? parseFloat(site.spec.properties.lng)  : 0,
    }
  });
  return sites;
}

async function SitesPage() {
  const sites = await getSites();  
  const params = {
    type: 'sites',
    menuItems: [],
    views: ['cards', 'table', 'map'],
    items: sites,
    refItems: [],
    columns: []
  }
  return (
    <div>
      <MultiView params={params} />
    </div>
  );
}

export default SitesPage;