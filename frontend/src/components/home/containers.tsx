import styled from '@emotion/styled'
import { StocksCard } from './card'

export const UserContainer = styled.div`
    display: flex;
    flex-direction: row;
    gap: 64px;
    align-items: center;
    justify-content: center;
`

export const UserTextContainer = styled.div`
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 32px;
`

export const AccountContainer = styled(UserTextContainer)`
    gap: 8px;
`

export const DataContainer = styled.div`
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 10px;
`

export const DataTextContainer = styled.div`
    display: flex;
    flex-direction: column;
    align-items: left;
    justify-content: center;
    gap: 10px;
`

export const DataValue = styled.div`
    font-family: Monsterrat-SemiBold;
    font-size: 20px;
`

export const DataName = styled.div`
    font-family: Monsterrat-Medium;
    font-size: 12px;
`

export const StickyContainer = styled.div`
    display: flex;
    flex-direction: column;
    position: sticky;
    top: 0;
    z-index: 10;
`
export const BottomContainer = styled.div`
    display: flex;
    flex-direction: row;
    gap: 32px;
`
export const TradesCardContainer = styled.div`
    display: flex;
    flex-direction: column;
    margin-top: 64px;
    gap: 32px;
    justify-content: flex-start;
    align-items: center;
`
export const TradesContainer = styled.div`
    display: flex;
    flex-direction: column;
    gap: 8px;
`
export const TradesComponentContainer = styled.div`
    width: 418px;
    display: flex;
    justify-content: space-between;
`
export const AddSellContainer = styled.div`
    display: flex;
    gap: 10px;
`
export const StocksCardContainer = styled.div`
    display: flex;
    flex-direction: column;
    margin-top: 64px;
    gap: 32px;
    justify-content: flex-start;
    align-items: center;
`
export const StocksContainer = styled(TradesContainer)`
    display: flex;
    flex-direction: column;
    gap: 8px;
`
export const StocksComponentContainer = styled(TradesComponentContainer)`
    width: 300px;
`

export const InputPopupContainer = styled.div`
    display: flex;
    flex-direction: row;
    gap: 16px;
`
