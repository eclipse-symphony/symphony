import React, { useState } from 'react';
import { getServerSession } from 'next-auth';
import { options } from '../../api/auth/[...nextauth]/options';
import CatalogEditor from '@/components/editors/CatalogEditor';
import {CatalogState, User} from '../../types';

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
    const catalogs = data
    .filter((catalog: CatalogState) => catalog.spec.catalogType === type);
    return catalogs;
  }

async function CatalogEditPage() {
    const [schemas] = await Promise.all([getCatalogs('schema')]);
    return <CatalogEditor schemas={schemas} />
}


export default CatalogEditPage;