import styled from '@emotion/styled'

export const BigBlackButton = styled.div`
    width: 498px;
    padding: 12px 24px;
    background-color: #000000;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    color: #ffffff;
    line-height: 29.26px;
    letter-spacing: 10%;
    &:hover {
        background-color: #374092;
    }
`

export const SmallBlackButton = styled(BigBlackButton)`
    width: 42px;
    height: 20px;
    padding: 0px 16px;
    font-family: Monsterrat-Medium;
    font-size: 16px;
    color: #ffffff;
`
