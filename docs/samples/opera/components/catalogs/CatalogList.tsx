import { CatalogState } from '../../app/types';
import CatalogCard from './CatalogCard';
interface CalalogListProps {
    catalogs: CatalogState[];
}
function CatalogList(props: CalalogListProps) {
    const { catalogs } = props;
    //create a map of catalogs
    const references: any = {};
    catalogs.forEach((catalog: CatalogState) => {
        references[catalog.spec.name] = catalog;
    });
    return (
        <div className='sitelist'>
            {catalogs.map((catalog: CatalogState) =>  
            <CatalogCard catalog={catalog} refCatalog={catalog.spec.parentName? references[catalog.spec.parentName]: null}/>)}
        </div>
    );
}
export default CatalogList;