import { CatalogVersionState } from '../../app/types';
import CatalogVersionCard from './CatalogVersionCard';
interface CalalogListProps {
    catalogversions: CatalogVersionState[];
}
function CatalogVersionList(props: CalalogListProps) {
    const { catalogversions } = props;
    //create a map of catalogversions
    const references: any = {};
    catalogversions.forEach((catalogversion: CatalogVersionState) => {
        references[catalogversion.spec.name] = catalogversion;
    });
    return (
        <div className='sitelist'>
            {catalogversions.map((catalogversion: CatalogVersionState) =>  
            <CatalogVersionCard catalogversion={catalogversion} refCatalogVersion={catalogversion.spec.parentName? references[catalogversion.spec.parentName]: null}/>)}
        </div>
    );
}
export default CatalogVersionList;