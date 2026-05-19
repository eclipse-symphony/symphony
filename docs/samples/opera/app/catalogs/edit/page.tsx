import React, { useState } from 'react';
import { getServerSession } from 'next-auth';
import { options } from '../../api/auth/[...nextauth]/options';
import CatalogVersionEditor from '@/components/editors/CatalogVersionEditor';
import {CatalogVersionState, User} from '../../types';

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
    const catalogversions = data
    .filter((catalogversion: CatalogVersionState) => catalogversion.spec.catalogType === type);
    return catalogversions;
  }

async function CatalogVersionEditPage() {
    const [schemas] = await Promise.all([getCatalogVersions('schema')]);
    return <CatalogVersionEditor schemas={schemas} />
}


export default CatalogVersionEditPage;