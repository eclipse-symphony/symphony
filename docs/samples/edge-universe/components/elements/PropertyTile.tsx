import React, { FC, useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCog } from '@fortawesome/free-solid-svg-icons';
import styles from './PropertyTile.module.css';
interface TitleProps {
  name: string;
  mode: string;
}

const PropertyTile: FC<TitleProps> = ({ name, value, mode }) => {
    const [tileMode, setTileMode] = useState<string>(mode);
    const [data, setData] = useState([]);
    
    const handleClick = () => {
        setTileMode((prevMode) => {
        if (prevMode === 'current') {
          fetchData()
          return 'history';
        }
        return 'current';
      });
    }
    
    const fetchData = async () => {
      //setData([{"hey":1}]);
    }   
  return (
    <div className={`flex flex-col rounded-xl bg-orange-500 p-5 m-5 ${tileMode === 'history' ? styles.history: styles.current}`}>
        <div className="flex justify-between items-center p-2 bg-gray-800 text-white">
            <p>{name}</p>
            <div className="flex">
            <i className="fas fa-plus mx-2"></i>
            <i className="fas fa-minus mx-2"></i>
            <FontAwesomeIcon icon={faCog} size="lg" className="mx-2" onClick={handleClick}/>
            </div>
        </div>
        <hr className="my-2 bg-gray-800"/>
        <div className="flex justify-center items-center my-5">          
            { tileMode === 'input'
              ? ( value.length == 2
                  ? <div className="text-5xl font-bold"><input className="w-4/12" value={`${value[0].data}`} /> / <input className="w-4/12" value={`${value[1].data}`} /></div>
                  : <div className="text-5xl font-bold"><input className="w-1/2" value={`${value[0].data}`} /></div>
              ): ( value.length == 2
                ? <p className="text-5xl font-bold">{value[0].data}/{value[1].data}</p>
                : <p className="text-5xl font-bold">{value[0].data}</p> 
              )
            }
            <p className="text-sm text-gray-800 ml-2">unit</p>
        </div>
    </div>
  );
};

export default PropertyTile;