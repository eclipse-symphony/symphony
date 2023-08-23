import { Catalog } from '../app/types';
import CatalogCard from './CatalogCard';
interface CalalogListProps {
    catalogs: Catalog[];
}
function CatalogList(props: CalalogListProps) {
    const { catalogs } = props;
    return (
        <div className='sitelist'>
            {catalogs.map((catalog: any) =>  <CatalogCard catalog={catalog} />)}
        </div>
    );
}
export default CatalogList;