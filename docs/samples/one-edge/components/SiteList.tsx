import { Site } from '../types';
import SiteCard from './SiteCard';
interface SiteListProps {
    sites: Site[];
}
function SiteList(props: SiteListProps) {
    const { sites } = props;
    return (
        <div className='sitelist'>
        {sites.map((site: any) =>  <SiteCard site={site} />)};
        </div>
    );
}
export default SiteList;