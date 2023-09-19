import { NextResponse } from "next/server"
import {CampaignState} from "../../../../types";
import { getServerSession } from 'next-auth';
import { options } from '../../../auth/[...nextauth]/options';
import { User } from '../../../../types';

export async function POST(
    request: Request,
    { params }: {
        params: {id: string }
    }
) {
    const body = await request.json();
    const campaignState: CampaignState = body;
    const session = await getServerSession(options);  
    const symphonyApi = process.env.SYMPHONY_API;
    const userObj: User | undefined = session?.user?? undefined;
    const res = await fetch( `${symphonyApi}activations/registry/${campaignState.id}`, {
        method: 'POST',
        headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
        },
        body: JSON.stringify({
            campaign: campaignState.id,
            name: campaignState.id,
            inputs: {}
        })
    });       
    // check if post is successful
    if (res.status !== 200) {
        return NextResponse.error();
    }
    return NextResponse.json({ });
}