import { CampaignState } from '../../app/types';
import CampaignCard from './CampaignCard';
interface CampaignCardListProps {
    campaigns: CampaignState[];
    activations?: any[];
}
function CampaignCardList(props: CampaignCardListProps) {
    const { campaigns, activations } = props;
    if (!campaigns) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>            

            {campaigns.map((campaign: any) =>  {
                const activation = activations?.find((activation: any) => activation.id === campaign.id);
                return <CampaignCard campaign={campaign} activation={activation}/>;
            })}
        </div>
    );
}
export default CampaignCardList;