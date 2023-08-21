import { CampaignState } from '../../types';
import CampaignCard from './CampaignCard';
interface CampaignCardListProps {
    campaigns: CampaignState[];
}
function CampaignCardList(props: CampaignCardListProps) {
    const { campaigns } = props;
    if (!campaigns) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>
            {campaigns.map((campaign: any) =>  <CampaignCard campaign={campaign} />)}
        </div>
    );
}
export default CampaignCardList;