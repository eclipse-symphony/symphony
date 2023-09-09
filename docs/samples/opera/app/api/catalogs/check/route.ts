import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from 'next-auth';
import { options } from '../../auth/[...nextauth]/options';
import { User, CatalogSpec } from '../../../types';
import { NextResponse } from "next/server"

export async function POST(
    request: Request)
{
    const body = await request.json();
    const catalog: CatalogSpec = body;
    const session = await getServerSession(options);  
    const symphonyApi = process.env.SYMPHONY_API;
    const userObj: User | undefined = session?.user?? undefined;
    const res = await fetch( `${symphonyApi}catalogs/check`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${userObj?.accessToken}`,
        },
        body: JSON.stringify(catalog)
    });       
    if (res.status == 200) {
        return NextResponse.json({ });
    }
    const responseBody = await res.json();
    return NextResponse.json(responseBody);
    // // check if post is successful
    // if (res.status !== 200) {
    //     return NextResponse.error();
    // }
    // const responseBody = await res.json();
    // return NextResponse.json(responseBody);
}