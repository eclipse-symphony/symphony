import MultiView from '@/components/MultiView';
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import { User } from '../types';

const getCampaignVersions = async () => {
    const session = await getServerSession(options);  
    const userObj: User | undefined = session?.user?? undefined;
    const symphonyApi = process.env.SYMPHONY_API;
    const res = await fetch( `${symphonyApi}campaignversions`, {
        method: 'GET',
        headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
        }
    });    
    const campaignversions = await res.json();
    return campaignversions;
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
async function CampaignVersionsPage() {
    const [campaignversions, activations] = await Promise.all([getCampaignVersions(), getActivations()]);  
    const params = {
        type: 'campaignversions',
        menuItems: [
            {
                name: 'Add CampaignVersion',
                href: '/campaignversions/add',                
            }
        ],
        views: ['cards', 'table'],
        items: campaignversions,
        refItems: activations,
        columns: []
    }
    return (
        <div>
            <MultiView params={params}  />
        </div>
    );
}

export default CampaignVersionsPage;