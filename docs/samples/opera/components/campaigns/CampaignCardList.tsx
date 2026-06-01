import { CampaignVersionState } from '../../app/types';
import CampaignVersionCard from './CampaignVersionCard';
interface CampaignVersionCardListProps {
    campaignversions: CampaignVersionState[];
    activations?: any[];
}
function CampaignVersionCardList(props: CampaignVersionCardListProps) {
    const { campaignversions, activations } = props;
    if (!campaignversions) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>            

            {campaignversions.map((campaignversion: any) =>  {
                const activation = activations?.find((activation: any) => activation.id === campaignversion.id);
                return <CampaignVersionCard campaignversion={campaignversion} activation={activation}/>;
            })}
        </div>
    );
}
export default CampaignVersionCardList;