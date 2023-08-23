import MultiView from '@/components/MultiView';
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
const getCampaigns = async () => {
    const session = await getServerSession(options);  
    console.log(session?.user?.accessToken);
    const symphonyApi = process.env.SYMPHONY_API;
    const res = await fetch( `${symphonyApi}campaigns`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${session?.user?.accessToken}`,
        }
    });    
    const campaigns = await res.json();
    return campaigns;
}
const getActivations = async () => {
    const session = await getServerSession(options);  
    console.log(session?.user?.accessToken);
    const symphonyApi = process.env.SYMPHONY_API;
    const res = await fetch( `${symphonyApi}activations/registry`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${session?.user?.accessToken}`,
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
    }
    return (
        <div>
            <MultiView params={params} />
        </div>
    );
}

export default CampaignsPage;