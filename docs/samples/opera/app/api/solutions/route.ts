import { NextResponse } from "next/server"
import { getServerSession } from 'next-auth';
import { options } from '../auth/[...nextauth]/options';
import { User } from '../../types';

export async function GET(
    request: Request,
    { params }: {
        params: {id: string }
    }
) {
    const session = await getServerSession(options);  
    const symphonyApi = process.env.SYMPHONY_API;
    const userObj: User | undefined = session?.user?? undefined;
    const res = await fetch( `${symphonyApi}solutions`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
        }
    });       
    // check if post is successful
    if (res.status !== 200) {
        return NextResponse.error();
    }
    return NextResponse.json(res.json());
}