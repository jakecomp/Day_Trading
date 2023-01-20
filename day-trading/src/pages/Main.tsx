import { Routes, Route } from 'react-router-dom'
import { SignIn } from './SignIn'
import { SignUp } from './SignUp'

const Main = () => (
    <Routes>
        <Route
            path='/'
            element={
                <>
                    <SignUp />
                </>
            }
        ></Route>
        <Route
            path='/signin'
            element={
                <>
                    <SignIn />
                </>
            }
        ></Route>
    </Routes>
)
export default Main
