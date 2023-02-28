import styled from '@emotion/styled'

export const StatusCard = styled.div`
    width: 1032px;
    height: 146px;
    background-color: #ffffff;
    box-shadow: 0 8px 24px 0 rgba(0, 0, 0, 0.05);
    border-radius: 16px;
    display: flex;
    flex-direction: row;
    gap: 64px;
    justify-content: center;
    align-items: center;
`

export const TradesCard = styled(StatusCard)`
    width: 655px;
    height: 637px;
    gap: 32px;
`

export const StocksCard = styled(TradesCard)`
    width: 345px;
    height: 637px;
`
