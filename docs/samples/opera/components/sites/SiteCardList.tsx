'use client';

import { Site } from '../../app/types';
import SiteCard from './SiteCard';

interface SiteCardListProps {
    sites: Site[];
}
function SiteCardList(props: SiteCardListProps) {
    const { sites } = props;
    return (
        <div className='sitelist'>
            {sites.map((site: any) =>  <SiteCard site={site} />)}
        </div>
    );
}
export default SiteCardList;