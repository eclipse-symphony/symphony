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
async function CampaignsPage() {
    const campaigns = await getCampaigns();  
    const params = {
        type: 'campaigns',
        menuItems: [
            {
                name: 'Add Campaign',
                href: '/campaigns/add',                
            }
        ],
        views: ['cards', 'table', 'map'],
        items: campaigns,
    }
    return (
        <div>
            <MultiView params={params} />
        </div>
    );
}

export default CampaignsPage;