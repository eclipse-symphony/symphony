import MultiView from '@/components/MultiView';
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import { User } from '../types';

const getCampaigns = async () => {
    const session = await getServerSession(options);  
    const userObj: User | undefined = session?.user?? undefined;
    const symphonyApi = process.env.SYMPHONY_API;
    const res = await fetch( `${symphonyApi}campaigns`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
        }
    });    
    const campaigns = await res.json();
    return campaigns;
}
const getActivations = async () => {
    const session = await getServerSession(options);  
    const userObj: User | undefined = session?.user?? undefined;
    const symphonyApi = process.env.SYMPHONY_API;
    const res = await fetch( `${symphonyApi}activations/registry`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
        }
    });    
    const activations = await res.json();
    return activations;
}
async function CampaignsPage() {
    const [campaigns, activations] = await Promise.all([getCampaigns(), getActivations()]);  
    const params = {
        type: 'campaigns',
        menuItems: [
            {
                name: 'Add Campaign',
                href: '/campaigns/add',                
            }
        ],
        views: ['cards', 'table'],
        items: campaigns,
        refItems: activations,
        columns: []
    }
    return (
        <div>
            <MultiView params={params}  />
        </div>
    );
}

export default CampaignsPage;