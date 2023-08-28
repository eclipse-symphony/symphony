import { CatalogState } from '../../app/types';
import AssetCard from './AssetCard';
interface AssetListProps {
    catalogs: CatalogState[];
}
function AssetList(props: AssetListProps) {
    const { catalogs } = props;
    //create a map of catalogs
    const references: any = {};
    catalogs.forEach((catalog: CatalogState) => {
        references[catalog.spec.name] = catalog;
    });
    return (
        <div className='sitelist'>
            {catalogs.map((catalog: CatalogState) =>  
            <AssetCard catalog={catalog} refCatalog={catalog.spec.metadata?.['override']? references[catalog.spec.metadata['override']]: null}/>)}
        </div>
    );
}
export default AssetList;